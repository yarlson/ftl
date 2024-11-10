package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/logs"
)

var follow bool

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Fetch logs from remote deployments",
	Long: `Fetch logs from the specified service running on remote servers.
Use the -f flag to stream logs in real-time.`,
	Args: cobra.ExactArgs(1),
	Run:  runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream logs in real-time")
}

func runLogs(cmd *cobra.Command, args []string) {
	serviceName := args[0]

	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		console.Error("Failed to parse config file:", err)
		return
	}

	if err := getLogsFromServers(cfg, serviceName, follow); err != nil {
		console.Error("Failed to fetch logs:", err)
		return
	}
}

func getLogsFromServers(cfg *config.Config, serviceName string, follow bool) error {
	for _, server := range cfg.Servers {
		console.Info(fmt.Sprintf("Fetching logs from server %s...", server.Host))

		runner, err := connectToServer(server)
		if err != nil {
			console.Error(fmt.Sprintf("Failed to connect to server %s: %v", server.Host, err))
			continue
		}
		defer runner.Close()

		logger := logs.NewLogger(runner)

		ctx := context.Background()
		if follow {
			err = logger.StreamLogs(ctx, cfg.Project.Name, serviceName)
		} else {
			err = logger.FetchLogs(ctx, cfg.Project.Name, serviceName)
		}

		if err != nil {
			console.Error(fmt.Sprintf("Failed to fetch logs from server %s: %v", server.Host, err))
			continue
		}
	}

	return nil
}
