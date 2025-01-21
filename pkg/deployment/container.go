package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/yarlson/ftl/pkg/config"
)

type containerInfo struct {
	ID     string
	Config struct {
		Image  string
		Env    []string
		Labels map[string]string
	}
	Image string
	State struct {
		Status string
	}
	NetworkSettings struct {
		Networks map[string]struct{ Aliases []string }
	}
	HostConfig struct {
		Binds []string
	}
}

type ContainerStatusType int

const (
	ContainerStatusRunning ContainerStatusType = iota
	ContainerStatusStopped
	ContainerStatusNotFound
	ContainerStatusError
)

func (d *Deployment) getContainerStatus(project, service string) (ContainerStatusType, error) {
	getContainerInfo, err := d.getContainerInfo(project, service)
	if err != nil {
		if strings.Contains(err.Error(), "no container found") {
			return ContainerStatusNotFound, nil
		}

		return ContainerStatusError, fmt.Errorf("failed to get container info: %w", err)
	}

	if getContainerInfo.State.Status != "running" {
		return ContainerStatusStopped, nil
	}

	return ContainerStatusRunning, nil
}

func (d *Deployment) getContainerID(project, service string) (string, error) {
	info, err := d.getContainerInfo(project, service)
	if err != nil {
		return "", err
	}

	return info.ID, err
}

func (d *Deployment) getContainerInfo(network, service string) (*containerInfo, error) {
	output, err := d.runCommand(context.Background(), "docker", "ps", "-aq", "--filter", fmt.Sprintf("network=%s", network))
	if err != nil {
		return nil, fmt.Errorf("failed to get container IDs: %w", err)
	}

	containerIDs := strings.Fields(output)
	for _, cid := range containerIDs {
		inspectOutput, err := d.runCommand(context.Background(), "docker", "inspect", cid)
		if err != nil {
			continue
		}

		var containerInfos []containerInfo
		if err := json.Unmarshal([]byte(inspectOutput), &containerInfos); err != nil || len(containerInfos) == 0 {
			continue
		}

		if aliases, ok := containerInfos[0].NetworkSettings.Networks[network]; ok {
			for _, alias := range aliases.Aliases {
				if alias == service {
					return &containerInfos[0], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no container found with alias %s in network %s", service, network)
}

func (d *Deployment) performHealthChecks(container string, healthCheck *config.ServiceHealthCheck) error {
	if healthCheck == nil {
		return nil
	}

	for i := 0; i < healthCheck.Retries; i++ {
		output, err := d.runCommand(context.Background(), "docker", "inspect", "--format={{.State.Health.Status}}", container)
		if err == nil && strings.TrimSpace(output) == "healthy" {
			return nil
		}
		time.Sleep(healthCheck.Interval)
	}

	output, err := d.runCommand(context.Background(), "docker", "logs", container)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %v", err)
	}

	lines := strings.Split(output, "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}
	trimmedOutput := strings.Join(lines, "\n")

	colorCodeRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	cleanedOutput := colorCodeRegex.ReplaceAllString(trimmedOutput, "")
	grayOutput := "\x1b[90m" + cleanedOutput + "\x1b[0m"

	return fmt.Errorf("container failed to become healthy\n\x1b[93mOutput from the container:\x1b[0m\n%s", grayOutput)
}

func (d *Deployment) startContainer(container string) error {
	_, err := d.runCommand(context.Background(), "docker", "start", container)
	if err != nil {
		return fmt.Errorf("failed to start container for %s: %v", container, err)
	}

	return nil
}

func (d *Deployment) createContainer(project string, service *config.Service, suffix string) error {
	container := containerName(project, service.Name, suffix)

	args := []string{"run"}

	if service.Container != nil && service.Container.RunOnce {
		args = append(args, "--rm")
	} else {
		args = append(args, "--detach")
	}

	args = append(args, []string{"--name", container, "--network", project, "--network-alias", service.Name + suffix, "--restart", "unless-stopped"}...)

	for _, value := range service.Env {
		args = append(args, "-e", value)
	}

	for _, volume := range service.Volumes {
		if unicode.IsLetter(rune(volume[0])) {
			volume = fmt.Sprintf("%s-%s", project, volume)
		}
		args = append(args, "-v", volume)
	}

	var healthCheckArgs []string

	if service.HealthCheck != nil {
		healthCheckArgs = []string{
			"--health-cmd", fmt.Sprintf("curl -sf http://localhost:%d%s || exit 1", service.Port, service.HealthCheck.Path),
			"--health-interval", fmt.Sprintf("%ds", int(service.HealthCheck.Interval.Seconds())),
			"--health-retries", fmt.Sprintf("%d", service.HealthCheck.Retries),
			"--health-timeout", fmt.Sprintf("%ds", int(service.HealthCheck.Timeout.Seconds())),
		}
	}

	if service.Container != nil && service.Container.HealthCheck != nil {
		healthCheckArgs = []string{
			"--health-cmd", service.Container.HealthCheck.Cmd,
			"--health-interval", service.Container.HealthCheck.Interval,
			"--health-retries", fmt.Sprintf("%d", service.Container.HealthCheck.Retries),
			"--health-timeout", service.Container.HealthCheck.Timeout,
			"--health-start-period", service.Container.HealthCheck.StartPeriod,
			"--health-start-timeout", service.Container.HealthCheck.StartTimeout,
		}
	}

	args = append(args, healthCheckArgs...)

	for _, port := range service.LocalPorts {
		args = append(args, "-p", fmt.Sprintf("127.0.0.1:%d:%d", port, port))
	}

	if len(service.Forwards) > 0 {
		for _, forward := range service.Forwards {
			args = append(args, "-p", forward)
		}
	}

	hash, err := service.Hash()
	if err != nil {
		return fmt.Errorf("failed to generate config hash: %w", err)
	}
	args = append(args, "--label", fmt.Sprintf("ftl.config-hash=%s", hash))

	if len(service.Entrypoint) > 0 {
		args = append(args, "--entrypoint", strings.Join(service.Entrypoint, " "))
	}

	image := service.Image
	if image == "" {
		image = fmt.Sprintf("%s-%s", project, service.Name)
	}
	args = append(args, image)

	if service.Command != "" {
		args = append(args, service.Command)
	}

	_, err = d.runCommand(context.Background(), "docker", args...)
	return err
}

func (d *Deployment) containerShouldBeUpdated(project string, service *config.Service) (bool, error) {
	containerInfo, err := d.getContainerInfo(project, service.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get container info: %w", err)
	}

	imageHash, err := d.getImageHash(service.Image)
	if err != nil {
		return false, fmt.Errorf("failed to get image hash: %w", err)
	}

	if service.Image == "" && service.ImageUpdated {
		return true, nil
	}

	if service.Image != "" && containerInfo.Image != imageHash {
		return true, nil
	}

	hash, err := service.Hash()
	if err != nil {
		return false, fmt.Errorf("failed to generate config hash: %w", err)
	}

	return containerInfo.Config.Labels["ftl.config-hash"] != hash, nil
}

func containerName(project, service, suffix string) string {
	return fmt.Sprintf("%s-%s%s", project, service, suffix)
}
