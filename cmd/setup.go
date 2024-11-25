package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/server"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare servers for deployment",
	Long: `Setup configures servers defined in ftl.yaml for deployment.
Run this once for each new server before deploying your application.`,
	Run: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	cfg, err := parseConfig("ftl.yaml")
	if err != nil {
		console.Error("Failed to parse config file:", err)
		return
	}

	dockerCreds, err := getDockerCredentials(cfg.Services)
	if err != nil {
		console.Error("Failed to get Docker credentials:", err)
		return
	}

	newUserPassword, err := getUserPassword()
	if err != nil {
		console.Error("Failed to read password:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	events := server.SetupServers(ctx, cfg, dockerCreds, newUserPassword)

	spinnerGroup := console.NewSpinnerGroup()
	defer spinnerGroup.StopAll()

	for event := range events {
		if err := spinnerGroup.HandleEvent(event); err != nil {
			console.Error("Setup failed:", err)
			return
		}
	}

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
