// cmd/build.go

package cmd

import (
	"context"
	"fmt"

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
	spinnerManager := console.NewSpinnerGroup(nil) // No MultiPrinter needed for single spinners

	if err := buildAndPushServices(ctx, cfg.Services, builder, skipPush, spinnerManager); err != nil {
		console.Error("Build process failed:", err)
		return
	}
}

func buildAndPushServices(ctx context.Context, services []config.Service, builder *build.Build, skipPush bool, spinnerGroup *console.SpinnerGroup) error {
	for _, service := range services {
		if err := buildAndPushService(ctx, service, builder, skipPush, spinnerGroup); err != nil {
			return fmt.Errorf("failed to build service %s: %w", service.Name, err)
		}
	}
	return nil
}

func buildAndPushService(ctx context.Context, service config.Service, builder *build.Build, skipPush bool, spinnerGroup *console.SpinnerGroup) error {
	buildMessage := fmt.Sprintf("Building service %s", service.Name)
	buildSuccessMessage := fmt.Sprintf("Service %s built successfully", service.Name)
	if err := spinnerGroup.RunWithSpinner(buildMessage, func() error {
		return builder.Build(ctx, service.Image, service.Path)
	}, buildSuccessMessage); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	if skipPush {
		return nil
	}

	pushMessage := fmt.Sprintf("Pushing service %s", service.Name)
	pushSuccessMessage := fmt.Sprintf("Service %s pushed successfully", service.Name)
	if err := spinnerGroup.RunWithSpinner(pushMessage, func() error {
		return builder.Push(ctx, service.Image)
	}, pushSuccessMessage); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}
