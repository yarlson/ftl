package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/tunnel"
)

var tunnelsCmd = &cobra.Command{
	Use:   "tunnels",
	Short: "Create SSH tunnels for dependencies",
	Long: `Create SSH tunnels for all dependencies defined in ftl.yaml,
forwarding local ports to remote ports.`,
	Run: runTunnels,
}

func init() {
	rootCmd.AddCommand(tunnelsCmd)
}

func runTunnels(cmd *cobra.Command, args []string) {
	sm := console.NewSpinnerManager()
	sm.Start()
	defer sm.Stop()

	spinner := sm.AddSpinner("tunnels", "Establishing SSH tunnels")

	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		spinner.ErrorWithMessagef("Failed to parse config file: %v", err)
		return
	}

	tunnels := tunnel.CollectDependencyTunnels(cfg)
	if len(tunnels) == 0 {
		spinner.ErrorWithMessage("No dependencies with ports found in the configuration.")
		return
	}

	// Use a cancelable context so we can shut down tunnels on Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tunnel.StartTunnels(
		ctx,
		cfg.Server.Host, cfg.Server.Port,
		cfg.Server.User, cfg.Server.SSHKey,
		tunnels,
	)
	if err != nil {
		spinner.ErrorWithMessagef("Failed to establish tunnels: %v", err)
		return
	}

	// If no error arrived in 2 seconds, we assume success (like the original):
	spinner.Complete()
	sm.Stop()

	console.Success("SSH tunnels established. Press Ctrl+C to exit.")

	// Same old signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	console.Info("Shutting down tunnels...")
	cancel()
	time.Sleep(1 * time.Second)
}
