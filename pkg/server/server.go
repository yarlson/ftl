package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gossh "golang.org/x/crypto/ssh"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/runner/remote"
	"github.com/yarlson/ftl/pkg/ssh"
)

// DockerCredentials holds the username and password for Docker authentication.
type DockerCredentials struct {
	Username string
	Password string
}

// SetupServers performs the setup on all servers concurrently.
func SetupServers(ctx context.Context, cfg *config.Config, dockerCreds DockerCredentials, newUserPassword string, sm *console.SpinnerManager) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(cfg.Servers))

	for _, server := range cfg.Servers {
		wg.Add(1)
		go func(server config.Server) {
			defer wg.Done()

			if err := setupServer(ctx, server, dockerCreds, newUserPassword, sm); err != nil {
				errChan <- fmt.Errorf("[%s] Setup failed: %w", server.Host, err)
				return
			}
		}(server)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during server setup: %v", errs)
	}

	return nil
}

func setupServer(ctx context.Context, cfg config.Server, dockerCreds DockerCredentials, newUserPassword string, sm *console.SpinnerManager) error {
	spinner := sm.AddSpinner("connecting", fmt.Sprintf("[%s] Connecting to server", cfg.Host))

	sshClient, rootKey, err := ssh.FindKeyAndConnectWithUser(cfg.Host, cfg.Port, "root", cfg.SSHKey)
	if err != nil {
		spinner.ErrorWithMessagef("Failed to connect via SSH: %v", err)
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer sshClient.Close()

	spinner.Complete()

	runner := remote.NewRunner(sshClient)
	cfg.RootSSHKey = string(rootKey)

	spinner = sm.AddSpinner("software", fmt.Sprintf("[%s] Installing software", cfg.Host))
	if err := installSoftware(ctx, runner); err != nil {
		spinner.ErrorWithMessagef("Failed to install software: %v", err)
		return fmt.Errorf("installing software: %w", err)
	}
	spinner.Complete()

	spinner = sm.AddSpinner("firewall", fmt.Sprintf("[%s] Configuring firewall", cfg.Host))
	if err := configureFirewall(ctx, runner); err != nil {
		spinner.ErrorWithMessagef("Failed to configure firewall: %v", err)
		return fmt.Errorf("configuring firewall: %w", err)
	}
	spinner.Complete()

	spinner = sm.AddSpinner("user", fmt.Sprintf("[%s] Creating user %s", cfg.Host, cfg.User))
	if err := createUser(ctx, runner, cfg.User, newUserPassword); err != nil {
		spinner.ErrorWithMessagef("Failed to create user: %v", err)
		return fmt.Errorf("creating user: %w", err)
	}
	spinner.Complete()

	spinner = sm.AddSpinner("sshkey", fmt.Sprintf("[%s] Setting up SSH key", cfg.Host))
	if err := setupSSHKey(ctx, runner, cfg); err != nil {
		spinner.ErrorWithMessagef("Failed to setup SSH key: %v", err)
		return fmt.Errorf("setting up SSH key: %w", err)
	}
	spinner.Complete()

	if dockerCreds.Username != "" && dockerCreds.Password != "" {
		spinner = sm.AddSpinner("docker", fmt.Sprintf("[%s] Logging into Docker Hub", cfg.Host))
		if err := dockerLogin(ctx, runner, dockerCreds); err != nil {
			spinner.ErrorWithMessagef("Failed to login to Docker: %v", err)
			return fmt.Errorf("docker login: %w", err)
		}
		spinner.Complete()
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

func setupSSHKey(ctx context.Context, runner *remote.Runner, server config.Server) error {
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
