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
