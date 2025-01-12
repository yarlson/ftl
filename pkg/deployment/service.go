package deployment

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yarlson/ftl/pkg/config"
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
		if err := d.startContainer(service); err != nil {
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

func (d *Deployment) startContainer(service *config.Service) error {
	_, err := d.runCommand(context.Background(), "docker", "start", service.Name)
	if err != nil {
		return fmt.Errorf("failed to start container for %s: %v", service.Name, err)
	}

	return nil
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
