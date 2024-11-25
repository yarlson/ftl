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
		for _, srvConfig := range cfg.Servers {
			wg.Add(1)
			go func(srvConfig config.Server) {
				defer wg.Done()
				hostname := srvConfig.Host

				if err := setupServer(ctx, srvConfig, dockerCreds, newUserPassword, events); err != nil {
					events <- console.Event{
						Type:    console.EventTypeError,
						Message: fmt.Sprintf("[%s] Setup failed: %v", hostname, err),
						Name:    hostname,
					}
					return
				}

				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Setup completed successfully", hostname),
					Name:    hostname,
				}
			}(srvConfig)
		}
		wg.Wait()
	}()
	return events
}

func setupServer(ctx context.Context, srvConfig config.Server, dockerCreds DockerCredentials, newUserPassword string, events chan<- console.Event) error {
	hostname := srvConfig.Host

	sshClient, rootKey, err := ssh.FindKeyAndConnectWithUser(srvConfig.Host, srvConfig.Port, "root", srvConfig.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer sshClient.Close()

	runner := remote.NewRunner(sshClient)
	srvConfig.RootSSHKey = string(rootKey)

	tasks := []struct {
		name   string
		action func() error
	}{
		{
			name: "Install Software",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Installing software", hostname),
					Name:    "install_software",
				}
				if err := installServerSoftware(ctx, runner); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Software installed", hostname),
					Name:    "install_software",
				}
				return nil
			},
		},
		{
			name: "Configure Firewall",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Configuring firewall", hostname),
					Name:    "configure_firewall",
				}
				if err := configureServerFirewall(ctx, runner); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] Firewall configured", hostname),
					Name:    "configure_firewall",
				}
				return nil
			},
		},
		{
			name: "Create User",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Creating user %s", hostname, srvConfig.User),
					Name:    "create_user",
				}
				if err := createServerUser(ctx, runner, srvConfig.User, newUserPassword); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] User %s created", hostname, srvConfig.User),
					Name:    "create_user",
				}
				return nil
			},
		},
		{
			name: "Setup SSH Key",
			action: func() error {
				events <- console.Event{
					Type:    console.EventTypeStart,
					Message: fmt.Sprintf("[%s] Setting up SSH key", hostname),
					Name:    "setup_ssh_key",
				}
				if err := setupServerSSHKey(ctx, runner, srvConfig); err != nil {
					return err
				}
				events <- console.Event{
					Type:    console.EventTypeFinish,
					Message: fmt.Sprintf("[%s] SSH key set up", hostname),
					Name:    "setup_ssh_key",
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
						Message: fmt.Sprintf("[%s] Logging into Docker Hub", hostname),
						Name:    "docker_login",
					}
					if err := dockerLogin(ctx, runner, dockerCreds); err != nil {
						return err
					}
					events <- console.Event{
						Type:    console.EventTypeFinish,
						Message: fmt.Sprintf("[%s] Logged into Docker Hub", hostname),
						Name:    "docker_login",
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

func installServerSoftware(ctx context.Context, runner *remote.Runner) error {
	commands := []string{
		"apt-get update",
		"apt-get install -y apt-transport-https ca-certificates curl wget git software-properties-common",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -",
		"add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" -y",
		"apt-get update",
		"apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin",
	}
	return runner.RunCommands(ctx, commands)
}

func configureServerFirewall(ctx context.Context, runner *remote.Runner) error {
	commands := []string{
		"apt-get install -y ufw",
		"ufw default deny incoming",
		"ufw default allow outgoing",
		"ufw allow 22/tcp",
		"ufw allow 80/tcp",
		"ufw allow 443/tcp",
		"echo 'y' | ufw enable",
	}
	return runner.RunCommands(ctx, commands)
}

func createServerUser(ctx context.Context, runner *remote.Runner, username, password string) error {
	checkUserCmd := fmt.Sprintf("id -u %s > /dev/null 2>&1", username)
	if _, err := runner.RunCommand(ctx, checkUserCmd); err == nil {
		return nil
	}

	commands := []string{
		fmt.Sprintf("adduser --gecos '' --disabled-password %s", username),
		fmt.Sprintf("echo '%s:%s' | chpasswd", username, password),
		fmt.Sprintf("usermod -aG docker %s", username),
	}
	return runner.RunCommands(ctx, commands)
}

func setupServerSSHKey(ctx context.Context, runner *remote.Runner, srvConfig config.Server) error {
	keyData, err := readSSHKey(srvConfig.SSHKey)
	if err != nil {
		return err
	}

	pubKey, err := parsePublicKey(keyData)
	if err != nil {
		return err
	}

	user := srvConfig.User
	sshDir := fmt.Sprintf("/home/%s/.ssh", user)
	authKeysFile := fmt.Sprintf("%s/authorized_keys", sshDir)

	commands := []string{
		fmt.Sprintf("mkdir -p %s", sshDir),
		fmt.Sprintf("echo '%s' | tee -a %s", pubKey, authKeysFile),
		fmt.Sprintf("chown -R %s:%s %s", user, user, sshDir),
		fmt.Sprintf("chmod 700 %s", sshDir),
		fmt.Sprintf("chmod 600 %s", authKeysFile),
	}
	return runner.RunCommands(ctx, commands)
}

func dockerLogin(ctx context.Context, runner *remote.Runner, creds DockerCredentials) error {
	command := fmt.Sprintf("docker login -u %s -p %s", creds.Username, creds.Password)
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
	privKey, err := gossh.ParsePrivateKey(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	return string(gossh.MarshalAuthorizedKey(privKey.PublicKey())), nil
}
