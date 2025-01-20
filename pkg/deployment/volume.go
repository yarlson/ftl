package deployment

import (
	"context"
	"fmt"
)

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
