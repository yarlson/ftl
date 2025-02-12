package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/yarlson/pin"

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
	pTunnel := pin.New("Establishing SSH tunnels", pin.WithSpinnerColor(pin.ColorCyan), pin.WithTextColor(pin.ColorYellow))
	cancelTunnel := pTunnel.Start(context.Background())

	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		pTunnel.Fail(fmt.Sprintf("Failed to parse config file: %v", err))
		cancelTunnel()
		return
	}

	tunnels := tunnel.CollectDependencyTunnels(cfg)
	if len(tunnels) == 0 {
		pTunnel.Fail("No dependencies with ports found in the configuration.")
		cancelTunnel()
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
		pTunnel.Fail(fmt.Sprintf("Failed to establish tunnels: %v", err))
		cancelTunnel()
		return
	}

	pTunnel.Stop("SSH tunnels established")
	cancelTunnel()

	console.Success("SSH tunnels established. Press Ctrl+C to exit.")

	// Same old signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	console.Info("Shutting down tunnels...")
	cancel()
	time.Sleep(1 * time.Second)
}
