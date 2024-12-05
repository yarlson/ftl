package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yarlson/ftl/pkg/config"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/ssh"
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

	tunnelsCmd.Flags().StringP("server", "s", "", "Server name or index to connect to")
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

	serverName, _ := cmd.Flags().GetString("server")
	serverConfig, err := selectServer(cfg, serverName)
	if err != nil {
		spinner.ErrorWithMessagef("Server selection failed: %v", err)
		return
	}

	user := serverConfig.User

	tunnels, err := collectDependencyTunnels(cfg)
	if err != nil {
		spinner.ErrorWithMessagef("Failed to collect dependencies: %v", err)
		return
	}
	if len(tunnels) == 0 {
		spinner.ErrorWithMessage("No dependencies with ports found in the configuration.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errorChan := make(chan error, len(tunnels))

	for _, tunnel := range tunnels {
		wg.Add(1)
		go func(t TunnelConfig) {
			defer wg.Done()
			err := ssh.CreateSSHTunnel(ctx, serverConfig.Host, serverConfig.Port, user, serverConfig.SSHKey, t.LocalPort, t.RemoteAddr)
			if err != nil {
				errorChan <- fmt.Errorf("Tunnel %s -> %s failed: %v", t.LocalPort, t.RemoteAddr, err)
			}
		}(tunnel)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	select {
	case err := <-errorChan:
		spinner.ErrorWithMessagef("Failed to establish tunnels: %v", err)
		return
	case <-time.After(2 * time.Second):
		spinner.Complete()
	}

	sm.Stop()

	console.Success("SSH tunnels established. Press Ctrl+C to exit.")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	console.Info("Shutting down tunnels...")
	cancel()
	time.Sleep(1 * time.Second)
}

func selectServer(cfg *config.Config, serverName string) (config.Server, error) {
	if serverName != "" {
		for _, srv := range cfg.Servers {
			if srv.Host == serverName || srv.User == serverName {
				return srv, nil
			}
		}
		return config.Server{}, fmt.Errorf("server not found in configuration: %s", serverName)
	} else if len(cfg.Servers) == 1 {
		return cfg.Servers[0], nil
	} else {
		return config.Server{}, fmt.Errorf("multiple servers defined. Please specify a server using the --server flag")
	}
}

type TunnelConfig struct {
	LocalPort  string
	RemoteAddr string
}

func collectDependencyTunnels(cfg *config.Config) ([]TunnelConfig, error) {
	var tunnels []TunnelConfig
	for _, dep := range cfg.Dependencies {
		for _, port := range dep.Ports {
			tunnels = append(tunnels, TunnelConfig{
				LocalPort:  fmt.Sprintf("%d", port),
				RemoteAddr: fmt.Sprintf("localhost:%d", port),
			})
		}
	}
	return tunnels, nil
}
