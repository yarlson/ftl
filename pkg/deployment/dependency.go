package deployment

import (
	"context"
	"fmt"
	"sync"

	"github.com/yarlson/ftl/pkg/config"
)

func (d *Deployment) deployDependencies(ctx context.Context, project string, dependencies []config.Dependency) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(dependencies))

	for _, dep := range dependencies {
		wg.Add(1)

		go func(dep config.Dependency) {
			defer wg.Done()

			if err := d.startDependency(project, &dep); err != nil {
				errChan <- fmt.Errorf("failed to deploy dependency %s: %w", dep.Name, err)
				return
			}
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

func (d *Deployment) startDependency(project string, dependency *config.Dependency) error {
	service := &config.Service{
		Name:       dependency.Name,
		Image:      dependency.Image,
		Volumes:    dependency.Volumes,
		Env:        dependency.Env,
		LocalPorts: dependency.Ports,
	}
	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to start container for %s: %v", dependency.Image, err)
	}

	return nil
}
