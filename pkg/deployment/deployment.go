package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/proxy"
)

const (
	newContainerSuffix = "_new"
)

type Runner interface {
	CopyFile(ctx context.Context, from, to string) error
	Host() string
	RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error)
}

type ImageSyncer interface {
	Sync(ctx context.Context, image string) (bool, error)
	CompareImages(ctx context.Context, image string) (bool, error)
}

type Deployment struct {
	runner Runner
	syncer ImageSyncer
	sm     *console.SpinnerManager
}

func NewDeployment(runner Runner, syncer ImageSyncer, sm *console.SpinnerManager) *Deployment {
	return &Deployment{runner: runner, syncer: syncer, sm: sm}
}

func (d *Deployment) Deploy(ctx context.Context, project string, cfg *config.Config) error {
	hostname := d.runner.Host()

	// Create project network
	spinner := d.sm.AddSpinner("network", fmt.Sprintf("[%s] Creating network...", hostname))
	if err := d.createNetwork(project); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to create network: %w", err)
	}
	spinner.Complete()

	// Create volumes
	if err := d.createVolumes(ctx, project, cfg.Volumes); err != nil {
		return fmt.Errorf("failed to create volumes: %w", err)
	}

	// Deploy dependencies
	if err := d.deployDependencies(ctx, project, cfg.Dependencies); err != nil {
		return fmt.Errorf("failed to deploy dependencies: %w", err)
	}

	// Deploy services
	if err := d.deployServices(ctx, project, cfg.Services); err != nil {
		return fmt.Errorf("failed to deploy services: %w", err)
	}

	// Setup proxy
	if err := d.startProxy(ctx, project, cfg); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	return nil
}

func (d *Deployment) createVolumes(ctx context.Context, project string, volumes []string) error {
	hostname := d.runner.Host()

	for _, volume := range volumes {
		spinner := d.sm.AddSpinner("volume", fmt.Sprintf("[%s] Creating volume %s", hostname, volume))

		if err := d.createVolume(ctx, project, volume); err != nil {
			spinner.ErrorWithMessagef("Failed to create volume %s: %v", volume, err)
			return fmt.Errorf("failed to create volume %s: %w", volume, err)
		}

		spinner.Complete()
	}

	return nil
}

func (d *Deployment) deployDependencies(ctx context.Context, project string, dependencies []config.Dependency) error {
	hostname := d.runner.Host()
	var wg sync.WaitGroup
	errChan := make(chan error, len(dependencies))

	for _, dep := range dependencies {
		wg.Add(1)
		go func(dep config.Dependency) {
			defer wg.Done()

			spinner := d.sm.AddSpinner("dependency", fmt.Sprintf("[%s] Deploying dependency %s", hostname, dep.Name))

			if err := d.startDependency(project, &dep); err != nil {
				spinner.ErrorWithMessagef("Failed to deploy dependency %s: %v", dep.Name, err)
				errChan <- fmt.Errorf("failed to deploy dependency %s: %w", dep.Name, err)
				return
			}

			spinner.Complete()
		}(dep)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during dependency deployment: %v", errs)
	}

	return nil
}

func (d *Deployment) deployServices(ctx context.Context, project string, services []config.Service) error {
	hostname := d.runner.Host()
	var wg sync.WaitGroup
	errChan := make(chan error, len(services))

	for _, service := range services {
		wg.Add(1)
		go func(service config.Service) {
			defer wg.Done()

			spinner := d.sm.AddSpinner(service.Name, fmt.Sprintf("[%s] Deploying service %s", hostname, service.Name))

			if err := d.deployService(project, &service); err != nil {
				spinner.ErrorWithMessagef("Failed to deploy service %s: %v", service.Name, err)
				errChan <- fmt.Errorf("failed to deploy service %s: %w", service.Name, err)
				return
			}

			spinner.Complete()
		}(service)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during service deployment: %v", errs)
	}

	return nil
}

func (d *Deployment) startProxy(ctx context.Context, project string, cfg *config.Config) error {
	hostname := d.runner.Host()

	// Prepare project folder
	projectPath, err := d.prepareProjectFolder(project)
	if err != nil {
		return fmt.Errorf("failed to prepare project folder: %w", err)
	}

	// Prepare nginx config
	spinner := d.sm.AddSpinner("config", fmt.Sprintf("[%s] Preparing Nginx configuration", hostname))
	configPath, err := d.prepareNginxConfig(cfg, projectPath)
	if err != nil {
		spinner.Error()
		return fmt.Errorf("failed to prepare nginx config: %w", err)
	}
	spinner.Complete()

	// Deploy proxy service
	spinner = d.sm.AddSpinner("proxy", fmt.Sprintf("[%s] Deploying proxy service", hostname))
	service := &config.Service{
		Name:  "proxy",
		Image: "yarlson/zero-nginx:latest",
		Port:  80,
		Volumes: []string{
			projectPath + "/:/etc/nginx/ssl",
			configPath + ":/etc/nginx/conf.d",
		},
		Env: []string{
			"DOMAIN=" + cfg.Project.Domain,
			"EMAIL=" + cfg.Project.Email,
		},
		Forwards: []string{
			"80:80",
			"443:443",
		},
		HealthCheck: &config.HealthCheck{
			Path:     "/",
			Interval: time.Second,
			Timeout:  time.Second,
			Retries:  30,
		},
		Recreate: true,
	}

	if err := d.deployService(project, service); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to deploy proxy service: %w", err)
	}
	spinner.Complete()

	// Reload nginx config
	spinner = d.sm.AddSpinner("nginx", fmt.Sprintf("[%s] Reloading Nginx configuration", hostname))
	if err := d.reloadNginxConfig(ctx); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to reload nginx config: %w", err)
	}
	spinner.Complete()

	// Deploy cert renewer
	spinner = d.sm.AddSpinner("certrenewer", fmt.Sprintf("[%s] Deploying certificate renewer", hostname))
	if err := d.deployCertRenewer(project, cfg); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to deploy certificate renewer: %w", err)
	}
	spinner.Complete()

	return nil
}

func (d *Deployment) startDependency(project string, dependency *config.Dependency) error {
	service := &config.Service{
		Name:    dependency.Name,
		Image:   dependency.Image,
		Volumes: dependency.Volumes,
		Env:     dependency.Env,
	}
	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to start container for %s: %v", dependency.Image, err)
	}

	return nil
}

func (d *Deployment) installService(project string, service *config.Service) error {
	if err := d.createContainer(project, service, ""); err != nil {
		return fmt.Errorf("failed to start container for %s: %v", service.Image, err)
	}

	svcName := service.Name

	if err := d.performHealthChecks(svcName, service.HealthCheck); err != nil {
		return fmt.Errorf("install failed for %s: container is unhealthy: %w", svcName, err)
	}

	return nil
}

func (d *Deployment) updateService(project string, service *config.Service) error {
	svcName := service.Name

	if service.Recreate {
		if err := d.recreateService(project, service); err != nil {
			return fmt.Errorf("failed to recreate service %s: %w", service.Name, err)
		}
		return nil
	}

	if err := d.createContainer(project, service, newContainerSuffix); err != nil {
		return fmt.Errorf("failed to start new container for %s: %v", svcName, err)
	}

	if err := d.performHealthChecks(svcName+newContainerSuffix, service.HealthCheck); err != nil {
		if _, err := d.runCommand(context.Background(), "docker", "rm", "-f", svcName+newContainerSuffix); err != nil {
			return fmt.Errorf("update failed for %s: new container is unhealthy and cleanup failed: %v", svcName, err)
		}
		return fmt.Errorf("update failed for %s: new container is unhealthy: %w", svcName, err)
	}

	oldContID, err := d.switchTraffic(project, svcName)
	if err != nil {
		return fmt.Errorf("failed to switch traffic for %s: %v", svcName, err)
	}

	if err := d.cleanup(oldContID, svcName); err != nil {
		return fmt.Errorf("failed to cleanup for %s: %v", svcName, err)
	}

	return nil
}

func (d *Deployment) recreateService(project string, service *config.Service) error {
	oldContID, err := d.getContainerID(project, service.Name)
	if err != nil {
		return fmt.Errorf("failed to get container ID for %s: %v", service.Name, err)
	}

	if _, err := d.runCommand(context.Background(), "docker", "stop", oldContID); err != nil {
		return fmt.Errorf("failed to stop old container for %s: %v", service.Name, err)
	}

	if _, err := d.runCommand(context.Background(), "docker", "rm", oldContID); err != nil {
		return fmt.Errorf("failed to remove old container for %s: %v", service.Name, err)
	}

	if err := d.createContainer(project, service, ""); err != nil {
		return fmt.Errorf("failed to start new container for %s: %v", service.Name, err)
	}

	if err := d.performHealthChecks(service.Name, service.HealthCheck); err != nil {
		if _, rmErr := d.runCommand(context.Background(), "docker", "rm", "-f", service.Name); rmErr != nil {
			return fmt.Errorf("recreation failed for %s: new container is unhealthy and cleanup failed: %v (original error: %w)", service.Name, rmErr, err)
		}
		return fmt.Errorf("recreation failed for %s: new container is unhealthy: %w", service.Name, err)
	}

	return nil
}

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

func (d *Deployment) createContainer(project string, service *config.Service, suffix string) error {
	svcName := service.Name

	args := []string{"run", "-d", "--name", svcName + suffix, "--network", project, "--network-alias", svcName + suffix, "--restart", "unless-stopped"}

	for _, value := range service.Env {
		args = append(args, "-e", value)
	}

	for _, volume := range service.Volumes {
		if unicode.IsLetter(rune(volume[0])) {
			volume = fmt.Sprintf("%s-%s", project, volume)
		}
		args = append(args, "-v", volume)
	}

	if service.HealthCheck != nil {
		args = append(args, "--health-cmd", fmt.Sprintf("curl -sf http://localhost:%d%s || exit 1", service.Port, service.HealthCheck.Path))
		args = append(args, "--health-interval", fmt.Sprintf("%ds", int(service.HealthCheck.Interval.Seconds())))
		args = append(args, "--health-retries", fmt.Sprintf("%d", service.HealthCheck.Retries))
		args = append(args, "--health-timeout", fmt.Sprintf("%ds", int(service.HealthCheck.Timeout.Seconds())))
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

func (d *Deployment) startContainer(service *config.Service) error {
	_, err := d.runCommand(context.Background(), "docker", "start", service.Name)
	if err != nil {
		return fmt.Errorf("failed to start container for %s: %v", service.Name, err)
	}

	return nil
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

	return fmt.Errorf("container failed to become healthy\n---\n%s\n---", output)
}

func (d *Deployment) switchTraffic(project, service string) (string, error) {
	newContainer := service + newContainerSuffix
	oldContainer, err := d.getContainerID(project, service)
	if err != nil {
		return "", fmt.Errorf("failed to get old container ID: %v", err)
	}

	cmds := [][]string{
		{"docker", "network", "disconnect", project, newContainer},
		{"docker", "network", "connect", "--alias", service, project, newContainer},
	}

	for _, cmd := range cmds {
		if _, err := d.runCommand(context.Background(), cmd[0], cmd[1:]...); err != nil {
			return "", fmt.Errorf("failed to execute command '%s': %v", strings.Join(cmd, " "), err)
		}
	}

	time.Sleep(1 * time.Second)

	cmds = [][]string{
		{"docker", "network", "disconnect", project, oldContainer},
	}

	for _, cmd := range cmds {
		if _, err := d.runCommand(context.Background(), cmd[0], cmd[1:]...); err != nil {
			return "", fmt.Errorf("failed to execute command '%s': %v", strings.Join(cmd, " "), err)
		}
	}

	return oldContainer, nil
}

func (d *Deployment) cleanup(oldContID, service string) error {
	cmds := [][]string{
		{"docker", "stop", oldContID},
		{"docker", "rm", oldContID},
		{"docker", "rename", service + newContainerSuffix, service},
	}

	for _, cmd := range cmds {
		if _, err := d.runCommand(context.Background(), cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("failed to execute command '%s': %v", strings.Join(cmd, " "), err)
		}
	}

	return nil
}

func (d *Deployment) pullImage(imageName string) (string, error) {
	_, err := d.runCommand(context.Background(), "docker", "pull", imageName)
	if err != nil {
		return "", err
	}

	output, err := d.runCommand(context.Background(), "docker", "images", "--no-trunc", "--format={{.ID}}", imageName)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func (d *Deployment) getImageHash(imageName string) (string, error) {
	output, err := d.runCommand(context.Background(), "docker", "inspect", "--format={{.Id}}", imageName)
	if err != nil {
		return "", err
	}

	if strings.Contains(output, "Error: No such ") {
		return "", nil
	}

	return strings.TrimSpace(output), nil
}

func (d *Deployment) runCommand(ctx context.Context, command string, args ...string) (string, error) {
	output, err := d.runner.RunCommand(ctx, command, args...)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	outputBytes, readErr := io.ReadAll(output)
	if readErr != nil {
		return "", fmt.Errorf("failed to read command output: %v (original error: %w)", readErr, err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func (d *Deployment) makeProjectFolder(projectName string) error {
	projectPath, err := d.projectFolder(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project folder path: %w", err)
	}

	_, err = d.runCommand(context.Background(), "mkdir", "-p", projectPath)
	return err
}

func (d *Deployment) projectFolder(projectName string) (string, error) {
	homeDir, err := d.runCommand(context.Background(), "sh", "-c", "echo $HOME")
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	homeDir = strings.TrimSpace(homeDir)
	projectPath := filepath.Join(homeDir, "projects", projectName)

	return projectPath, nil
}

func (d *Deployment) prepareProjectFolder(project string) (string, error) {
	if err := d.makeProjectFolder(project); err != nil {
		return "", fmt.Errorf("failed to create project folder: %w", err)
	}

	return d.projectFolder(project)
}

func (d *Deployment) prepareNginxConfig(cfg *config.Config, projectPath string) (string, error) {
	nginxConfig, err := proxy.GenerateNginxConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to generate nginx config: %w", err)
	}

	nginxConfig = strings.TrimSpace(nginxConfig)

	configPath := filepath.Join(projectPath, "nginx")
	_, err = d.runCommand(context.Background(), "mkdir", "-p", configPath)
	if err != nil {
		return "", fmt.Errorf("failed to create nginx config directory: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "nginx-config-*.conf")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(nginxConfig); err != nil {
		return "", fmt.Errorf("failed to write nginx config to temporary file: %w", err)
	}

	return configPath, d.runner.CopyFile(context.Background(), tmpFile.Name(), filepath.Join(configPath, "default.conf"))
}

func (d *Deployment) deployService(project string, service *config.Service) error {
	err := d.updateImage(project, service)
	if err != nil {
		return err
	}

	containerStatus, err := d.getContainerStatus(project, service.Name)
	if err != nil {
		return err
	}

	if containerStatus == ContainerStatusNotFound {
		if err := d.installService(project, service); err != nil {
			return fmt.Errorf("failed to install service %s: %w", service.Name, err)
		}

		return nil
	}

	containerShouldBeUpdated, err := d.containerShouldBeUpdated(project, service)
	if err != nil {
		return err
	}

	if containerShouldBeUpdated {
		if err := d.updateService(project, service); err != nil {
			return fmt.Errorf("failed to update service %s due to image change: %w", service.Name, err)
		}

		return nil
	}

	if containerStatus == ContainerStatusStopped {
		if err := d.startContainer(service); err != nil {
			return fmt.Errorf("failed to start container %s: %w", service.Name, err)
		}

		return nil
	}

	return nil
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

func (d *Deployment) updateImage(project string, service *config.Service) error {
	if service.Image == "" {
		updated, err := d.syncer.Sync(context.Background(), fmt.Sprintf("%s-%s", project, service.Name))
		if err != nil {
			return err
		}
		service.ImageUpdated = updated
	}

	_, err := d.pullImage(service.Image)
	if err != nil {
		return err
	}

	return nil
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

func (d *Deployment) networkExists(network string) (bool, error) {
	output, err := d.runCommand(context.Background(), "docker", "network", "ls", "--format", "{{.Name}}")
	if err != nil {
		return false, fmt.Errorf("failed to list Docker networks: %w", err)
	}

	networks := strings.Split(strings.TrimSpace(output), "\n")
	for _, n := range networks {
		if strings.TrimSpace(n) == network {
			return true, nil
		}
	}
	return false, nil
}

func (d *Deployment) createNetwork(network string) error {
	exists, err := d.networkExists(network)
	if err != nil {
		return fmt.Errorf("failed to check if network exists: %w", err)
	}

	if exists {
		return nil
	}

	_, err = d.runCommand(context.Background(), "docker", "network", "create", network)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}

func (d *Deployment) createVolume(ctx context.Context, project, volume string) error {
	volumeName := fmt.Sprintf("%s-%s", project, volume)
	if _, err := d.runCommand(context.Background(), "docker", "volume", "inspect", volumeName); err == nil {
		return nil
	}

	_, err := d.runCommand(ctx, "docker", "volume", "create", volumeName)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	return nil
}

func (d *Deployment) reloadNginxConfig(ctx context.Context) error {
	_, err := d.runCommand(ctx, "docker", "exec", "proxy", "nginx", "-s", "reload")
	return err
}

func (d *Deployment) deployCertRenewer(project string, cfg *config.Config) error {
	service := &config.Service{
		Name:  "certrenewer",
		Image: "yarlson/zero-nginx:1.27-alpine3.19-zero0.2.0-0.2",
		Volumes: []string{
			"certs:/etc/nginx/ssl",
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		Env: []string{
			"DOMAIN=" + cfg.Project.Domain,
			"EMAIL=" + cfg.Project.Email,
			"PROXY_CONTAINER_NAME=proxy",
		},
		Entrypoint: []string{"/renew-certificates.sh"},
		Recreate:   true,
	}

	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to deploy certrenewer service: %w", err)
	}

	return nil
}
