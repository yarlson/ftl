package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/logs"
)

var (
	follow bool
	tail   int
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Fetch logs from remote deployment",
	Long: `Fetch logs from the specified service running on remote server.
If no service is specified, logs from all services will be fetched.
Use the -f flag to stream logs in real-time.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream logs in real-time")
	logsCmd.Flags().IntVarP(&tail, "tail", "n", -1, "Number of lines to show from the end of the logs")
}

func runLogs(cmd *cobra.Command, args []string) {
	var serviceName string
	if len(args) > 0 {
		serviceName = args[0]
	}

	if follow && !cmd.Flags().Lookup("tail").Changed {
		tail = 100
	}

	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		console.Error("Failed to parse config file:", err)
		return
	}

	if err := getLogs(cfg, serviceName, follow, tail); err != nil {
		console.Error("Failed to fetch logs:", err)
		return
	}
}

func getLogs(cfg *config.Config, serviceName string, follow bool, tail int) error {
	services := []string{}

	if serviceName != "" {
		services = append(services, serviceName)
	} else {
		for _, service := range cfg.Services {
			services = append(services, service.Name)
		}
	}

	console.Info(fmt.Sprintf("Fetching logs from server %s...", cfg.Server.Host))

	runner, err := connectToServer(cfg.Server)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %v", cfg.Server.Host, err)
	}

	defer func() {
		_ = runner.Close()
	}()

	logger := logs.NewLogger(runner)
	ctx := context.Background()

	if err := logger.FetchLogs(ctx, cfg.Project.Name, services, follow, tail); err != nil {
		return fmt.Errorf("failed to fetch logs from server %s: %v", cfg.Server.Host, err)
	}

	return nil
}
