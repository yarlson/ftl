package deployment

import (
	"context"
	"fmt"
	"github.com/yarlson/ftl/pkg/runner/local"
	"io"
	"path/filepath"
	"strings"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
)

const (
	newContainerSuffix = "_new"
)

type Runner interface {
	CopyFile(ctx context.Context, from, to string) error
	Host() string
	RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error)
}

type ImageSyncer interface {
	Sync(ctx context.Context, image string) (bool, error)
	CompareImages(ctx context.Context, image string) (bool, error)
}

type Deployment struct {
	runner      Runner
	localRunner *local.Runner
	syncer      ImageSyncer
	sm          *console.SpinnerManager
}

func NewDeployment(runner Runner, syncer ImageSyncer, sm *console.SpinnerManager) *Deployment {
	return &Deployment{runner: runner, syncer: syncer, sm: sm, localRunner: local.NewRunner()}
}

func (d *Deployment) Deploy(ctx context.Context, project string, cfg *config.Config) error {
	hostname := d.runner.Host()

	// Create project network
	spinner := d.sm.AddSpinner("network", fmt.Sprintf("[%s] Creating network...", hostname))
	if err := d.createNetwork(project); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to create network: %w", err)
	}
	spinner.Complete()

	// Create volumes
	cfg.Volumes = append(cfg.Volumes, "certs")
	if err := d.createVolumes(ctx, project, cfg.Volumes); err != nil {
		return fmt.Errorf("failed to create volumes: %w", err)
	}

	// Deploy dependencies
	if err := d.deployDependencies(ctx, project, cfg.Dependencies); err != nil {
		return fmt.Errorf("failed to deploy dependencies: %w", err)
	}

	// Start tunnels if needed
	tunnelCtx, tunnelCancel := context.WithCancel(ctx)
	defer tunnelCancel()

	if hasLocalHooks(cfg) {
		if err := d.startTunnels(tunnelCtx, cfg); err != nil {
			return fmt.Errorf("failed to start tunnels: %w", err)
		}
	}

	// Deploy services
	if err := d.deployServices(ctx, project, cfg.Services); err != nil {
		return fmt.Errorf("failed to deploy services: %w", err)
	}

	tunnelCancel()

	// Setup proxy
	if err := d.startProxy(ctx, project, cfg); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	return nil
}

func (d *Deployment) runCommand(ctx context.Context, command string, args ...string) (string, error) {
	output, err := d.runner.RunCommand(ctx, command, args...)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	outputBytes, readErr := io.ReadAll(output)
	if readErr != nil {
		return "", fmt.Errorf("failed to read command output: %v (original error: %w)", readErr, err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func (d *Deployment) runLocalCommand(ctx context.Context, command string, args ...string) (string, error) {
	output, err := d.runner.RunCommand(ctx, command, args...)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	outputBytes, readErr := io.ReadAll(output)
	if readErr != nil {
		return "", fmt.Errorf("failed to read command output: %v (original error: %w)", readErr, err)
	}

	return strings.TrimSpace(string(outputBytes)), nil
}

func (d *Deployment) makeProjectFolder(projectName string) error {
	projectPath, err := d.projectFolder(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project folder path: %w", err)
	}

	_, err = d.runCommand(context.Background(), "mkdir", "-p", projectPath)
	return err
}

func (d *Deployment) projectFolder(projectName string) (string, error) {
	homeDir, err := d.runCommand(context.Background(), "sh", "-c", "echo $HOME")
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	homeDir = strings.TrimSpace(homeDir)
	projectPath := filepath.Join(homeDir, "projects", projectName)

	return projectPath, nil
}
