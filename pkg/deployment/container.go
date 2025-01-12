package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

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

func (d *Deployment) performHealthChecks(container string, healthCheck *config.HealthCheck) error {
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
