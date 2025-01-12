package deployment

import (
	"context"
	"fmt"
	"strings"

	"github.com/yarlson/ftl/pkg/config"
)

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