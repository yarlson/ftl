package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/build"
	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/executor/local"
)

// buildCmd represents the build command
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
		console.ErrPrintln("Failed to parse config file:", err)
		return
	}

	skipPush, err := cmd.Flags().GetBool("skip-push")
	if err != nil {
		console.ErrPrintln("Failed to get skip-push flag:", err)
		return
	}

	executor := local.NewExecutor()
	builder := build.NewBuild(executor)

	ctx := context.Background()

	if err := buildAndPushServices(ctx, cfg.Services, builder, skipPush); err != nil {
		console.ErrPrintln("Build process failed:", err)
		return
	}

	message := "Build process completed successfully."
	if skipPush {
		message += " Images were not pushed due to --skip-push flag."
	}
	console.Success(message)
}

func buildAndPushServices(ctx context.Context, services []config.Service, builder *build.Build, skipPush bool) error {
	for _, service := range services {
		if err := buildAndPushService(ctx, service, builder, skipPush); err != nil {
			return fmt.Errorf("failed to build service %s: %w", service.Name, err)
		}
	}
	return nil
}

func buildAndPushService(ctx context.Context, service config.Service, builder *build.Build, skipPush bool) error {
	if err := builder.Build(ctx, service.Image, service.Path); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	if skipPush {
		return nil
	}

	if err := builder.Push(ctx, service.Image); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}
