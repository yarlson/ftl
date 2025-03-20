package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/yarlson/ftl/pkg/config"
)

// ContainerDetails holds information from a Docker inspect.
type ContainerDetails struct {
	ID     string
	Config struct {
		Image  string
		Env    []string
		Labels map[string]string
	}
	Image           string
	State           struct{ Status string }
	NetworkSettings struct {
		Networks map[string]struct{ Aliases []string }
	}
	HostConfig struct{ Binds []string }
}

// ContainerStatus represents the status of a container.
type ContainerStatus int

const (
	ContainerStatusRunning ContainerStatus = iota
	ContainerStatusStopped
	ContainerStatusNotFound
	ContainerStatusError
)

// CommandRunner defines an interface for running commands.
type CommandRunner interface {
	RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error)
}

// DockerManager manages Docker containers.
type DockerManager struct {
	runner CommandRunner
}

// NewDockerManager creates a new DockerManager.
func NewDockerManager(runner CommandRunner) *DockerManager {
	return &DockerManager{runner: runner}
}

// GetContainerStatus returns the status of a container identified by networkName and serviceName.
func (dm *DockerManager) GetContainerStatus(networkName, serviceName string) (ContainerStatus, error) {
	details, err := dm.findContainerDetails(networkName, serviceName)
	if err != nil {
		if strings.Contains(err.Error(), "no container found") {
			return ContainerStatusNotFound, nil
		}
		return ContainerStatusError, fmt.Errorf("failed to get container details: %w", err)
	}

	if details.State.Status != "running" {
		return ContainerStatusStopped, nil
	}

	return ContainerStatusRunning, nil
}

// GetContainerID returns the container ID for the given network and service.
func (dm *DockerManager) GetContainerID(networkName, serviceName string) (string, error) {
	details, err := dm.findContainerDetails(networkName, serviceName)
	if err != nil {
		return "", err
	}
	return details.ID, nil
}

// findContainerDetails retrieves Docker inspect information for the container matching
// the given networkName and serviceName alias.
func (dm *DockerManager) findContainerDetails(networkName, serviceName string) (*ContainerDetails, error) {
	output, err := dm.runCommand(context.Background(), "docker", "ps", "-aq", "--filter", fmt.Sprintf("network=%s", networkName))
	if err != nil {
		return nil, fmt.Errorf("failed to get container IDs: %w", err)
	}

	containerIDs := strings.Fields(output)
	for _, containerID := range containerIDs {
		inspectOutput, err := dm.runCommand(context.Background(), "docker", "inspect", containerID)
		if err != nil {
			continue
		}

		var containers []ContainerDetails
		if err := json.Unmarshal([]byte(inspectOutput), &containers); err != nil || len(containers) == 0 {
			continue
		}

		if networkConfig, ok := containers[0].NetworkSettings.Networks[networkName]; ok {
			for _, alias := range networkConfig.Aliases {
				if alias == serviceName {
					return &containers[0], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no container found with alias %s in network %s", serviceName, networkName)
}

// CheckContainerHealth performs health checks for the container with the given ID.
func (dm *DockerManager) CheckContainerHealth(containerID string, hc *config.ServiceHealthCheck) error {
	if hc == nil {
		return nil
	}

	for i := 0; i < hc.Retries; i++ {
		output, err := dm.runCommand(context.Background(), "docker", "inspect", "--format={{.State.Health.Status}}", containerID)
		if err == nil && strings.TrimSpace(output) == "healthy" {
			return nil
		}
		time.Sleep(hc.Interval)
	}

	output, err := dm.runCommand(context.Background(), "docker", "logs", containerID)
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

// StartContainer starts the container with the given ID.
func (dm *DockerManager) StartContainer(containerID string) error {
	_, err := dm.runCommand(context.Background(), "docker", "start", containerID)
	if err != nil {
		return fmt.Errorf("failed to start container %s: %v", containerID, err)
	}
	return nil
}

// CreateAndRunContainer creates and starts a container for the given service on the specified network.
func (dm *DockerManager) CreateAndRunContainer(networkName string, svc *config.Service, suffix string) error {
	containerName := generateContainerName(networkName, svc.Name, suffix)

	args := []string{"run"}
	if svc.Container != nil && svc.Container.RunOnce {
		args = append(args, "--rm")
	} else {
		args = append(args, "--detach")
	}

	args = append(args, []string{
		"--name", containerName,
		"--network", networkName,
		"--network-alias", svc.Name + suffix,
		"--restart", "unless-stopped",
	}...)

	for _, envVal := range svc.Env {
		args = append(args, "-e", envVal)
	}

	for _, vol := range svc.Volumes {
		if unicode.IsLetter(rune(vol[0])) {
			vol = fmt.Sprintf("%s-%s", networkName, vol)
		}
		args = append(args, "-v", vol)
	}

	var healthArgs []string
	if svc.HealthCheck != nil {
		healthArgs = []string{
			"--health-cmd", fmt.Sprintf("curl -sf http://localhost:%d%s || exit 1", svc.Port, svc.HealthCheck.Path),
			"--health-interval", fmt.Sprintf("%ds", int(svc.HealthCheck.Interval.Seconds())),
			"--health-retries", fmt.Sprintf("%d", svc.HealthCheck.Retries),
			"--health-timeout", fmt.Sprintf("%ds", int(svc.HealthCheck.Timeout.Seconds())),
		}
	}
	if svc.Container != nil && svc.Container.HealthCheck != nil {
		healthArgs = []string{
			"--health-cmd", svc.Container.HealthCheck.Cmd,
			"--health-interval", svc.Container.HealthCheck.Interval,
			"--health-retries", fmt.Sprintf("%d", svc.Container.HealthCheck.Retries),
		}
		if svc.Container.HealthCheck.Timeout != "" {
			healthArgs = append(healthArgs, "--health-timeout", svc.Container.HealthCheck.Timeout)
		}
		if svc.Container.HealthCheck.StartPeriod != "" {
			healthArgs = append(healthArgs, "--health-start-period", svc.Container.HealthCheck.StartPeriod)
		}
		if svc.Container.HealthCheck.StartTimeout != "" {
			healthArgs = append(healthArgs, "--health-start-timeout", svc.Container.HealthCheck.StartTimeout)
		}
	}
	args = append(args, healthArgs...)

	for _, port := range svc.LocalPorts {
		args = append(args, "-p", fmt.Sprintf("127.0.0.1:%d:%d", port, port))
	}
	for _, forward := range svc.Forwards {
		args = append(args, "-p", forward)
	}

	hash, err := svc.Hash()
	if err != nil {
		return fmt.Errorf("failed to generate config hash: %w", err)
	}
	args = append(args, "--label", fmt.Sprintf("ftl.config-hash=%s", hash))

	if len(svc.Entrypoint) > 0 {
		args = append(args, "--entrypoint", strings.Join(svc.Entrypoint, " "))
	}

	image := svc.Image
	if image == "" {
		image = fmt.Sprintf("%s-%s", networkName, svc.Name)
	}
	args = append(args, image)

	if svc.Command != "" {
		args = append(args, svc.Command)
	}
	if len(svc.CommandSlice) > 0 {
		args = append(args, svc.CommandSlice...)
	}

	_, err = dm.runCommand(context.Background(), "docker", args...)
	return err
}

// ContainerNeedsUpdate determines if a container should be updated based on its configuration and image.
func (dm *DockerManager) ContainerNeedsUpdate(networkName string, svc *config.Service) (bool, error) {
	details, err := dm.findContainerDetails(networkName, svc.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get container details: %w", err)
	}

	imageID, err := dm.fetchImageID(svc.Image)
	if err != nil {
		return false, fmt.Errorf("failed to get image ID: %w", err)
	}

	if svc.Image == "" && svc.ImageUpdated {
		return true, nil
	}
	if svc.Image != "" && details.Image != imageID {
		return true, nil
	}

	configHash, err := svc.Hash()
	if err != nil {
		return false, fmt.Errorf("failed to generate config hash: %w", err)
	}
	return details.Config.Labels["ftl.config-hash"] != configHash, nil
}

// generateContainerName constructs a container name using the networkName, serviceName, and suffix.
func generateContainerName(networkName, serviceName, suffix string) string {
	return fmt.Sprintf("%s-%s%s", networkName, serviceName, suffix)
}

// runCommand executes a command and returns its trimmed output.
func (dm *DockerManager) runCommand(ctx context.Context, command string, args ...string) (string, error) {
	output, err := dm.runner.RunCommand(ctx, command, args...)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}
	outputBytes, readErr := io.ReadAll(output)
	if readErr != nil {
		return "", fmt.Errorf("failed to read command output: %v (original error: %w)", readErr, err)
	}
	return strings.TrimSpace(string(outputBytes)), nil
}

// fetchImageID returns the image ID for the specified image.
func (dm *DockerManager) fetchImageID(imageName string) (string, error) {
	output, err := dm.runCommand(context.Background(), "docker", "inspect", "--format={{.Id}}", imageName)
	if err != nil {
		return "", err
	}
	if strings.Contains(output, "Error: No such ") {
		return "", nil
	}
	return strings.TrimSpace(output), nil
}

// PullImage pulls the specified image from the Docker registry and verifies it.
func (dm *DockerManager) PullImage(imageName string) error {
	_, err := dm.runCommand(context.Background(), "docker", "pull", imageName)
	if err != nil {
		return err
	}
	_, err = dm.runCommand(context.Background(), "docker", "images", "--no-trunc", "--format={{.ID}}", imageName)
	if err != nil {
		return err
	}
	return nil
}

// networkExists returns true if a Docker network with the specified name exists.
func (dm *DockerManager) networkExists(networkName string) (bool, error) {
	output, err := dm.runCommand(context.Background(), "docker", "network", "ls", "--format", "{{.Name}}")
	if err != nil {
		return false, fmt.Errorf("failed to list Docker networks: %w", err)
	}

	networks := strings.Split(strings.TrimSpace(output), "\n")
	for _, n := range networks {
		if strings.TrimSpace(n) == networkName {
			return true, nil
		}
	}
	return false, nil
}

// EnsureNetwork checks if the Docker network exists, and if not, creates it.
func (dm *DockerManager) EnsureNetwork(networkName string) error {
	exists, err := dm.networkExists(networkName)
	if err != nil {
		return fmt.Errorf("failed to check if network exists: %w", err)
	}

	if exists {
		return nil
	}

	_, err = dm.runCommand(context.Background(), "docker", "network", "create", networkName)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}

// CreateVolume creates a Docker volume for the specified project and volume name if it does not already exist.
func (dm *DockerManager) CreateVolume(ctx context.Context, project, volume string) error {
	volumeName := fmt.Sprintf("%s-%s", project, volume)
	if _, err := dm.runCommand(ctx, "docker", "volume", "inspect", volumeName); err == nil {
		return nil
	}

	_, err := dm.runCommand(ctx, "docker", "volume", "create", volumeName)
	if err != nil {
		return fmt.Errorf("failed to create volume %s: %w", volumeName, err)
	}

	return nil
}
