package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yarlson/ftl/pkg/imagesync"

	"github.com/yarlson/ftl/pkg/ssh"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/deployment"
	"github.com/yarlson/ftl/pkg/runner/remote"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application to configured servers",
	Long: `Deploy your application to all servers defined in ftl.yaml.
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

	if err := deployToServers(cfg); err != nil {
		console.Error("Deployment failed:", err)
		return
	}
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

func deployToServers(cfg *config.Config) error {
	for _, server := range cfg.Servers {
		if err := deployToServer(cfg.Project.Name, cfg, server); err != nil {
			return fmt.Errorf("failed to deploy to server %s: %w", server.Host, err)
		}
		console.Success(fmt.Sprintf("Successfully deployed to server %s", server.Host))
	}

	return nil
}

func deployToServer(project string, cfg *config.Config, server config.Server) error {
	console.Info(fmt.Sprintf("Deploying to server %s...", server.Host))

	runner, err := connectToServer(server)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer runner.Close()

	syncer := imagesync.NewImageSync(imagesync.Config{}, runner)
	deploy := deployment.NewDeployment(runner, syncer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := deploy.Deploy(ctx, project, cfg)

	multi := pterm.DefaultMultiPrinter

	spinners := make(map[string]*pterm.SpinnerPrinter)

	_, _ = multi.Start()
	defer func() { _, _ = multi.Stop() }()

	for event := range events {
		switch event.Type {
		case deployment.EventTypeStart:
			spinner := console.NewSpinnerWithWriter(event.Message, multi.NewWriter())
			spinners[event.Name] = spinner
		case deployment.EventTypeProgress:
			if spinner, ok := spinners[event.Name]; ok {
				spinner.UpdateText(event.Message)
			} else {
				console.Info(event.Message)
			}
		case deployment.EventTypeFinish:
			if spinner, ok := spinners[event.Name]; ok {
				spinner.Success(event.Message)
				delete(spinners, event.Name)
			} else {
				console.Success(event.Message)
			}
		case deployment.EventTypeError:
			if spinner, ok := spinners[event.Name]; ok {
				spinner.Fail(fmt.Sprintf("Deployment error: %s", event.Message))
				delete(spinners, event.Name)
			} else {
				console.Error(fmt.Sprintf("Deployment error: %s", event.Message))
			}
			return fmt.Errorf("deployment error: %s", event.Message)
		case deployment.EventTypeComplete:
			if spinner, ok := spinners[event.Name]; ok {
				spinner.Success(event.Message)
				delete(spinners, event.Name)
			} else {
				console.Success(event.Message)
			}
		default:
			if spinner, ok := spinners[event.Name]; ok {
				spinner.UpdateText(event.Message)
			} else {
				console.Info(event.Message)
			}
		}
	}

	for _, spinner := range spinners {
		_ = spinner.Stop()
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
