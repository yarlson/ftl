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

// SetupServers performs the setup on all servers concurrently and returns a channel of events.
func SetupServers(ctx context.Context, cfg *config.Config, dockerCreds DockerCredentials, newUserPassword string) <-chan console.Event {
	events := make(chan console.Event)

	go func() {
		defer close(events)
		var wg sync.WaitGroup

		for _, server := range cfg.Servers {
			wg.Add(1)
			go func(server config.Server) {
				defer wg.Done()
				host := server.Host

				if err := setupServer(ctx, server, dockerCreds, newUserPassword, events); err != nil {
					events <- console.Event{
						Type:    console.EventTypeError,
						Message: fmt.Sprintf("[%s] Setup failed: %v", host, err),
						Name:    host,
					}
					return
				}

				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Setup completed successfully", host),
					Name:    host,
				}
			}(server)
		}
		wg.Wait()
	}()

	return events
}

func setupServer(ctx context.Context, server config.Server, dockerCreds DockerCredentials, newUserPassword string, events chan<- console.Event) error {
	host := server.Host

	sshClient, rootKey, err := ssh.FindKeyAndConnectWithUser(server.Host, server.Port, "root", server.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer sshClient.Close()

	runner := remote.NewRunner(sshClient)
	server.RootSSHKey = string(rootKey)

	tasks := []struct {
		name   string
		action func() error
	}{
		{
			name: "Install Software",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Installing software", host),
					Name:    "installSoftware",
				}
				if err := installSoftware(ctx, runner); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Software installed", host),
					Name:    "installSoftware",
				}
				return nil
			},
		},
		{
			name: "Configure Firewall",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Configuring firewall", host),
					Name:    "configureFirewall",
				}
				if err := configureFirewall(ctx, runner); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Firewall configured", host),
					Name:    "configureFirewall",
				}
				return nil
			},
		},
		{
			name: "Create User",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Creating user %s", host, server.User),
					Name:    "createUser",
				}
				if err := createUser(ctx, runner, server.User, newUserPassword); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] User %s created", host, server.User),
					Name:    "createUser",
				}
				return nil
			},
		},
		{
			name: "Setup SSH Key",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Setting up SSH key", host),
					Name:    "setupSSHKey",
				}
				if err := setupSSHKey(ctx, runner, server); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] SSH key set up", host),
					Name:    "setupSSHKey",
				}
				return nil
			},
		},
		{
			name: "Docker Login",
			action: func() error {
				if dockerCreds.Username != "" && dockerCreds.Password != "" {
					events <- console.Event{
						Type:    console.EventTypeStart,
						Message: fmt.Sprintf("[%s] Logging into Docker Hub", host),
						Name:    "dockerLogin",
					}
					if err := dockerLogin(ctx, runner, dockerCreds); err != nil {
						return err
					}
					events <- console.Event{
						Type:    console.EventTypeFinish,
						Message: fmt.Sprintf("[%s] Logged into Docker Hub", host),
						Name:    "dockerLogin",
					}
				}
				return nil
			},
		},
	}

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task.action(); err != nil {
				return fmt.Errorf("failed during task %s: %w", task.name, err)
			}
		}
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
