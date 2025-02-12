package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/deployment"
	"github.com/yarlson/ftl/pkg/imagesync"
	"github.com/yarlson/ftl/pkg/runner/remote"
	"github.com/yarlson/ftl/pkg/ssh"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application to configured server",
	Long: `Deploy your application to the server defined in ftl.yaml.
This command handles the entire deployment process, ensuring
zero-downtime updates of your services.`,
	Run: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) {
	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		console.Error("Failed to parse config file:", err)
		return
	}

	if err := deployToServer(cfg.Project.Name, cfg, cfg.Server); err != nil {
		console.Error("Deployment failed:", err)
		return
	}

	console.Success("Deployment completed successfully")
}

func parseConfig(filename string) (*config.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg, err := config.ParseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

func deployToServer(project string, cfg *config.Config, server config.Server) error {
	hostname := server.Host

	// Connect to server
	runner, err := connectToServer(server)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %w", hostname, err)
	}
	defer runner.Close()

	// Create temp directory for docker sync
	localStore, err := os.MkdirTemp("", "dockersync-local")
	if err != nil {
		return fmt.Errorf("failed to create local store: %w", err)
	}

	// Initialize image syncer and deployment
	syncer := imagesync.NewImageSync(imagesync.Config{
		LocalStore:  localStore,
		MaxParallel: 1,
	}, runner)
	deploy := deployment.NewDeployment(runner, syncer)

	// Start deployment
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := deploy.Deploy(ctx, project, cfg); err != nil {
		return err
	}

	return nil
}

func connectToServer(server config.Server) (*remote.Runner, error) {
	sshKeyPath := filepath.Join(os.Getenv("HOME"), ".ssh", filepath.Base(server.SSHKey))
	sshClient, _, err := ssh.FindKeyAndConnectWithUser(server.Host, server.Port, server.User, sshKeyPath)
	if err != nil {
		return nil, err
	}

	return remote.NewRunner(sshClient), nil
}
