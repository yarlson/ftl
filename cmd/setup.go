package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yarlson/pin"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/server"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare server for deployment",
	Long: `Setup configures server defined in ftl.yaml for deployment.
Run this once for each new server before deploying your application.`,
	Run: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	pConfig := pin.New("Parsing configuration", pin.WithSpinnerColor(pin.ColorCyan))
	cancelConfig := pConfig.Start(context.Background())
	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		pConfig.Fail(fmt.Sprintf("Failed to parse config file: %v", err))
		cancelConfig()
		return
	}
	pConfig.Stop("Configuration parsed")
	cancelConfig()

	pDocker := pin.New("Checking Docker credentials", pin.WithSpinnerColor(pin.ColorCyan))
	cancelDocker := pDocker.Start(context.Background())
	dockerCreds, err := getDockerCredentials(cfg.Services)
	if err != nil {
		pDocker.Fail(fmt.Sprintf("Failed to get Docker credentials: %v", err))
		cancelDocker()
		return
	}
	pDocker.Stop("Docker credentials obtained")
	cancelDocker()

	newUserPassword, err := getUserPassword()
	if err != nil {
		console.Error("Failed to read password:", err)
		return
	}
	console.ClearPreviousLine()
	console.Success("Password set successfully")

	pSetup := pin.New("Setting up server", pin.WithSpinnerColor(pin.ColorCyan))
	cancelSetup := pSetup.Start(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := server.Setup(ctx, cfg, server.DockerCredentials{
		Username: dockerCreds.Username,
		Password: dockerCreds.Password,
	}, newUserPassword); err != nil {
		pSetup.Fail(fmt.Sprintf("Setup failed: %v", err))
		cancelSetup()
		return
	}
	pSetup.Stop("Server setup completed successfully")
	cancelSetup()

	console.Success("Server setup completed successfully.")
}

func getDockerCredentials(services []config.Service) (server.DockerCredentials, error) {
	var creds server.DockerCredentials

	if !needDockerHubLogin(services) {
		return creds, nil
	}

	console.Input("Enter Docker Hub username:")
	username, err := console.ReadLine()
	if err != nil {
		return creds, fmt.Errorf("failed to read Docker Hub username: %w", err)
	}

	console.Input("Enter Docker Hub password:")
	password, err := console.ReadPassword()
	if err != nil {
		return creds, fmt.Errorf("failed to read Docker Hub password: %w", err)
	}
	fmt.Println()

	return server.DockerCredentials{Username: username, Password: password}, nil
}

func getUserPassword() (string, error) {
	console.Input("Enter new user password:")
	password, err := console.ReadPassword()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	return password, nil
}

func needDockerHubLogin(services []config.Service) bool {
	for _, service := range services {
		if imageFromDockerHub(service.Image) {
			return true
		}
	}
	return false
}

func imageFromDockerHub(image string) bool {
	if image == "" {
		return false
	}
	parts := strings.SplitN(image, "/", 2)
	return len(parts) == 1 || !strings.Contains(parts[0], ".")
}
