package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gossh "golang.org/x/crypto/ssh"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/runner/remote"
	"github.com/yarlson/ftl/pkg/ssh"
	"github.com/yarlson/pin"
)

// DockerCredentials holds the username and password for Docker authentication.
type DockerCredentials struct {
	Username string
	Password string
}

// Setup performs the server setup with progress updates.
func Setup(ctx context.Context, cfg *config.Config, dockerCreds DockerCredentials, newUserPassword string, spinner *pin.Pin) error {
	spinner.UpdateMessage("Starting server setup on " + cfg.Server.Host + "...")
	if err := setupServer(ctx, cfg.Server, dockerCreds, newUserPassword, spinner); err != nil {
		return fmt.Errorf("[%s] Setup failed: %w", cfg.Server.Host, err)
	}
	spinner.UpdateMessage("Server setup completed successfully.")
	return nil
}

func setupServer(ctx context.Context, cfg *config.Server, dockerCreds DockerCredentials, newUserPassword string, spinner *pin.Pin) error {
	spinner.UpdateMessage("Establishing SSH connection to server " + cfg.Host + " as root...")
	sshClient, rootKey, err := ssh.FindKeyAndConnectWithUser(cfg.Host, cfg.Port, "root", cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	spinner.UpdateMessage("SSH connection established.")
	defer sshClient.Close()

	runner := remote.NewRunner(sshClient)
	cfg.RootSSHKey = string(rootKey)

	spinner.UpdateMessage("Installing required software...")
	if err := installSoftware(ctx, runner); err != nil {
		return fmt.Errorf("installing software: %w", err)
	}
	spinner.UpdateMessage("Software installation complete.")

	spinner.UpdateMessage("Configuring firewall...")
	if err := configureFirewall(ctx, runner); err != nil {
		return fmt.Errorf("configuring firewall: %w", err)
	}
	spinner.UpdateMessage("Firewall configuration complete.")

	spinner.UpdateMessage("Creating user account " + cfg.User + "...")
	if err := createUser(ctx, runner, cfg.User, newUserPassword); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	spinner.UpdateMessage("User account created.")

	spinner.UpdateMessage("Setting up SSH key for user " + cfg.User + "...")
	if err := setupSSHKey(ctx, runner, cfg); err != nil {
		return fmt.Errorf("setting up SSH key: %w", err)
	}
	spinner.UpdateMessage("SSH key setup complete.")

	if dockerCreds.Username != "" && dockerCreds.Password != "" {
		spinner.UpdateMessage("Logging into Docker registry...")
		if err := dockerLogin(ctx, runner, dockerCreds); err != nil {
			return fmt.Errorf("docker login: %w", err)
		}
		spinner.UpdateMessage("Docker login successful.")
	}

	return nil
}

func installSoftware(ctx context.Context, runner *remote.Runner) error {
	commands := []string{
		"apt-get update",
		"apt-get install -y apt-transport-https ca-certificates curl wget git software-properties-common",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -",
		`add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" -y`,
		"apt-get update",
		"apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin",
	}
	return runner.RunCommands(ctx, commands)
}

func configureFirewall(ctx context.Context, runner *remote.Runner) error {
	commands := []string{
		"apt-get install -y ufw",
		"ufw default deny incoming",
		"ufw default allow outgoing",
		"ufw allow 22/tcp",
		"ufw allow 80/tcp",
		"ufw allow 443/tcp",
		`echo "y" | ufw enable`,
	}
	return runner.RunCommands(ctx, commands)
}

func createUser(ctx context.Context, runner *remote.Runner, user, password string) error {
	checkUserCmd := fmt.Sprintf("id -u %s", user)
	if _, err := runner.RunCommand(ctx, checkUserCmd); err == nil {
		// User already exists
		return nil
	}

	commands := []string{
		fmt.Sprintf("adduser --gecos '' --disabled-password %s", user),
		fmt.Sprintf("echo '%s:%s' | chpasswd", user, password),
		fmt.Sprintf("usermod -aG docker %s", user),
	}
	return runner.RunCommands(ctx, commands)
}

func setupSSHKey(ctx context.Context, runner *remote.Runner, server *config.Server) error {
	keyData, err := readSSHKey(server.SSHKey)
	if err != nil {
		return err
	}

	publicKey, err := parsePublicKey(keyData)
	if err != nil {
		return err
	}

	user := server.User
	sshDir := fmt.Sprintf("/home/%s/.ssh", user)
	authKeysFile := filepath.Join(sshDir, "authorized_keys")

	commands := []string{
		fmt.Sprintf("mkdir -p %s", sshDir),
		fmt.Sprintf("echo '%s' | tee -a %s", publicKey, authKeysFile),
		fmt.Sprintf("chown -R %s:%s %s", user, user, sshDir),
		fmt.Sprintf("chmod 700 %s", sshDir),
		fmt.Sprintf("chmod 600 %s", authKeysFile),
	}
	return runner.RunCommands(ctx, commands)
}

func dockerLogin(ctx context.Context, runner *remote.Runner, creds DockerCredentials) error {
	command := fmt.Sprintf("echo '%s' | docker login -u %s --password-stdin", creds.Password, creds.Username)
	_, err := runner.RunCommand(ctx, command)
	return err
}

func readSSHKey(keyPath string) ([]byte, error) {
	if strings.HasPrefix(keyPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		keyPath = filepath.Join(homeDir, keyPath[1:])
	}
	return os.ReadFile(keyPath)
}

func parsePublicKey(keyData []byte) (string, error) {
	privateKey, err := gossh.ParsePrivateKey(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	return string(gossh.MarshalAuthorizedKey(privateKey.PublicKey())), nil
}
