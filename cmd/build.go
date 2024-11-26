package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/build"
	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/runner/local"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build your application Docker images",
	Long: `Build your application Docker images as defined in ftl.yaml.
This command handles the entire build process, including
building and pushing the Docker images to the registry.`,
	Run: runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().Bool("skip-push", false, "Skip pushing images to registry after building")
}

func runBuild(cmd *cobra.Command, args []string) {
	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		console.Error("Failed to parse config file:", err)
		return
	}

	skipPush, err := cmd.Flags().GetBool("skip-push")
	if err != nil {
		console.Error("Failed to get skip-push flag:", err)
		return
	}

	runner := local.NewRunner()
	builder := build.NewBuild(runner)

	ctx := context.Background()

	if err := buildAndPushServices(ctx, cfg.Project.Name, cfg.Services, builder, skipPush); err != nil {
		console.Error("Build process failed:", err)
		return
	}
}

// buildAndPushServices builds and pushes all services concurrently.
func buildAndPushServices(ctx context.Context, project string, services []config.Service, builder *build.Build, skipPush bool) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(services))
	eventChan := make(chan console.Event)
	done := make(chan struct{})

	spinnerGroup := console.NewGroup()
	defer spinnerGroup.StopAll()

	go func() {
		for event := range eventChan {
			if err := spinnerGroup.HandleEvent(event); err != nil {
				errChan <- err
			}
		}
		close(done)
	}()

	for _, svc := range services {
		wg.Add(1)
		go func(svc config.Service) {
			defer wg.Done()
			serviceName := svc.Name
			image := svc.Image
			if image == "" {
				image = fmt.Sprintf("%s-%s", project, serviceName)
			}

			eventChan <- console.Event{
				Type:    console.EventStart,
				Message: fmt.Sprintf("Building service %s", serviceName),
				Name:    serviceName,
			}
			if err := builder.Build(ctx, image, svc.Path); err != nil {
				eventChan <- console.Event{
					Type:    console.EventError,
					Message: fmt.Sprintf("Failed to build service %s: %v", serviceName, err),
					Name:    serviceName,
				}
				errChan <- fmt.Errorf("failed to build service %s: %w", serviceName, err)
				return
			}
			eventChan <- console.Event{
				Type:    console.EventFinish,
				Message: fmt.Sprintf("Service %s built successfully", serviceName),
				Name:    serviceName,
			}

			if skipPush || svc.Image == "" {
				return
			}

			eventChan <- console.Event{
				Type:    console.EventStart,
				Message: fmt.Sprintf("Pushing service %s", serviceName),
				Name:    serviceName,
			}
			if err := builder.Push(ctx, svc.Image); err != nil {
				eventChan <- console.Event{
					Type:    console.EventError,
					Message: fmt.Sprintf("Failed to push service %s: %v", serviceName, err),
					Name:    serviceName,
				}
				errChan <- fmt.Errorf("failed to push service %s: %w", serviceName, err)
				return
			}
			eventChan <- console.Event{
				Type:    console.EventFinish,
				Message: fmt.Sprintf("Service %s pushed successfully", serviceName),
				Name:    serviceName,
			}
		}(svc)
	}

	wg.Wait()
	close(eventChan)
	<-done
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during build/push: %v", errs)
	}

	return nil
}
