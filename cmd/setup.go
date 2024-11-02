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

// setupCmd represents the setup command
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
		console.ErrPrintln("Failed to parse config file:", err)
		return
	}

	dockerCreds, err := getDockerCredentials(cfg.Services)
	if err != nil {
		console.ErrPrintln("Failed to get Docker credentials:", err)
		return
	}

	newUserPassword, err := getUserPassword()
	if err != nil {
		console.ErrPrintln("Failed to read password:", err)
		return
	}

	if dockerCreds.Username != "" && dockerCreds.Password != "" {
		if err := server.DockerLogin(context.Background(), dockerCreds.Username, dockerCreds.Password); err != nil {
			console.ErrPrintln("Failed to login to Docker Hub:", err)
			return
		}
	}

	for _, s := range cfg.Servers {
		if err := setupServer(s, dockerCreds, newUserPassword); err != nil {
			console.ErrPrintln(fmt.Sprintf("Failed to setup server %s:", s.Host), err)
			continue
		}
		console.Success(fmt.Sprintf("Successfully set up server %s", s.Host))
	}

	console.Success("Server setup completed successfully.")
}

type dockerCredentials struct {
	Username string
	Password string
}

func getDockerCredentials(services []config.Service) (dockerCredentials, error) {
	var creds dockerCredentials

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

	return dockerCredentials{Username: username, Password: password}, nil
}

func getUserPassword() (string, error) {
	console.Input("Enter server user password:")
	password, err := console.ReadPassword()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	return password, nil
}

func setupServer(serverConfig config.Server, dockerCreds dockerCredentials, newUserPassword string) error {
	console.Info(fmt.Sprintf("Setting up server %s...", serverConfig.Host))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	s, err := server.NewServer(&serverConfig)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return s.RunSetup(ctx, dockerCreds.Username, dockerCreds.Password)
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
	parts := strings.SplitN(image, "/", 2)
	return len(parts) == 1 || !strings.Contains(parts[0], ".")
}
