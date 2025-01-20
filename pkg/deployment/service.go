package deployment

import (
	"context"
	"fmt"
	"github.com/yarlson/ftl/pkg/config"
	"strings"
	"sync"
	"time"
)

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
		container := containerName(project, service.Name, "")
		if err := d.startContainer(container); err != nil {
			return fmt.Errorf("failed to start container %s: %w", service.Name, err)
		}
		return nil
	}

	return nil
}

func (d *Deployment) installService(project string, service *config.Service) error {
	if err := d.createContainer(project, service, ""); err != nil {
		return fmt.Errorf("failed to start container for %s: %v", service.Image, err)
	}

	container := containerName(project, service.Name, "")

	if err := d.performHealthChecks(container, service.HealthCheck); err != nil {
		return fmt.Errorf("install failed for %s: container is unhealthy: %w", container, err)
	}

	if service.Hooks != nil && service.Hooks.Pre != nil && service.Hooks.Pre.Remote != "" {
		if err := d.runRemoteHook(context.Background(), container, service.Hooks.Pre.Remote); err != nil {
			return fmt.Errorf("remote pre-hook failed: %w", err)
		}
	}

	if service.Hooks != nil && service.Hooks.Post != nil && service.Hooks.Post.Remote != "" {
		if err := d.runRemoteHook(context.Background(), container, service.Hooks.Post.Remote); err != nil {
			return fmt.Errorf("remote pre-hook failed: %w", err)
		}
	}

	return nil
}

func (d *Deployment) updateService(project string, service *config.Service) error {
	container := containerName(project, service.Name, "")

	if service.Recreate {
		if err := d.recreateService(project, service); err != nil {
			return fmt.Errorf("failed to recreate service %s: %w", service.Name, err)
		}
		return nil
	}

	if err := d.createContainer(project, service, newContainerSuffix); err != nil {
		return fmt.Errorf("failed to start new container for %s: %v", container, err)
	}

	if err := d.performHealthChecks(container+newContainerSuffix, service.HealthCheck); err != nil {
		if _, err := d.runCommand(context.Background(), "docker", "rm", "-f", container+newContainerSuffix); err != nil {
			return fmt.Errorf("update failed for %s: new container is unhealthy and cleanup failed: %v", container, err)
		}
		return fmt.Errorf("update failed for %s: new container is unhealthy: %w", container, err)
	}

	if service.Hooks != nil && service.Hooks.Pre != nil && service.Hooks.Pre.Remote != "" {
		if err := d.runRemoteHook(context.Background(), container+newContainerSuffix, service.Hooks.Pre.Remote); err != nil {
			return fmt.Errorf("remote pre-hook failed: %w", err)
		}
	}

	oldContID, err := d.switchTraffic(project, service.Name)
	if err != nil {
		return fmt.Errorf("failed to switch traffic for %s: %v", container, err)
	}

	if err := d.cleanup(project, oldContID, service.Name); err != nil {
		return fmt.Errorf("failed to cleanup for %s: %v", container, err)
	}

	if service.Hooks != nil && service.Hooks.Post != nil && service.Hooks.Post.Remote != "" {
		if err := d.runRemoteHook(context.Background(), container, service.Hooks.Post.Remote); err != nil {
			return fmt.Errorf("remote pre-hook failed: %w", err)
		}
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

func (d *Deployment) switchTraffic(project, service string) (string, error) {
	newContainer := containerName(project, service, newContainerSuffix)
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

func (d *Deployment) cleanup(project, oldContID, service string) error {
	oldContainer := containerName(project, service, newContainerSuffix)
	newContainer := containerName(project, service, "")
	cmds := [][]string{
		{"docker", "stop", oldContID},
		{"docker", "rm", oldContID},
		{"docker", "rename", oldContainer, newContainer},
	}

	for _, cmd := range cmds {
		if _, err := d.runCommand(context.Background(), cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("failed to execute command '%s': %v", strings.Join(cmd, " "), err)
		}
	}

	return nil
}

// runRemoteHook executes the given command inside the specified container
func (d *Deployment) runRemoteHook(ctx context.Context, containerName, command string) error {
	if command == "" {
		return nil
	}

	dockerCmd := fmt.Sprintf("docker exec %s sh -c \"%s\"", containerName, command)

	if _, err := d.runCommand(ctx, "sh", "-c", dockerCmd); err != nil {
		return fmt.Errorf("failed to run remote hook in container %s: %w", containerName, err)
	}

	return nil
}
