package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/console"
	"github.com/yarlson/ftl/pkg/executor/local"
	sshPkg "github.com/yarlson/ftl/pkg/executor/ssh"
)

func DockerLogin(ctx context.Context, dockerUsername, dockerPassword string) error {
	executor := local.NewExecutor()

	if err := executor.RunCommandWithProgress(
		ctx,
		"Logging into Docker Hub...",
		"Logged into Docker Hub successfully.",
		[]string{
			fmt.Sprintf("docker login -u %s -p %s", dockerUsername, dockerPassword),
		},
	); err != nil {
		return fmt.Errorf("failed to configure docker hub: %w", err)
	}

	return nil
}

type Server struct {
	config *config.Server
	client *sshPkg.Client
}

func NewServer(config *config.Server) (*Server, error) {
	client, rootKey, err := sshPkg.FindKeyAndConnectWithUser(config.Host, config.Port, "root", config.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("failed to find a suitable SSH key and connect to the server: %w", err)
	}

	config.RootSSHKey = string(rootKey)

	return &Server{
		config: config,
		client: client,
	}, nil
}

func (s *Server) RunSetup(ctx context.Context, dockerUsername, dockerPassword string) error {
	console.Success("SSH connection to the server established.")

	if err := s.installServerSoftware(ctx); err != nil {
		return err
	}

	if err := s.configureServerFirewall(ctx); err != nil {
		return err
	}

	if err := s.createServerUser(ctx); err != nil {
		return err
	}

	if err := s.setupServerSSHKey(ctx); err != nil {
		return err
	}

	if dockerUsername != "" && dockerPassword != "" {
		if err := s.configureDockerHub(ctx, dockerUsername, dockerPassword); err != nil {
			return fmt.Errorf("failed to configure docker hub: %w", err)
		}
	}

	return nil
}

func (s *Server) installServerSoftware(ctx context.Context) error {
	commands := []string{
		"apt-get update",
		"apt-get install -y apt-transport-https ca-certificates curl wget git software-properties-common",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -",
		"add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" -y",
		"apt-get update",
		"apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin",
	}

	return s.client.RunCommandWithProgress(
		ctx,
		"Provisioning server with essential software...",
		"Essential software and Docker installed successfully.",
		commands,
	)
}

func (s *Server) configureServerFirewall(ctx context.Context) error {
	commands := []string{
		"apt-get install -y ufw",
		"ufw default deny incoming",
		"ufw default allow outgoing",
		"ufw allow 22/tcp",
		"ufw allow 80/tcp",
		"ufw allow 443/tcp",
		"echo 'y' | ufw enable",
	}

	return s.client.RunCommandWithProgress(
		ctx,
		"Configuring server firewall...",
		"Server firewall configured successfully.",
		commands,
	)
}

func (s *Server) createServerUser(ctx context.Context) error {
	checkUserCmd := fmt.Sprintf("id -u %s > /dev/null 2>&1", s.config.User)
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.client.RunCommand(checkCtx, checkUserCmd)
	if err == nil {
		console.Warning(fmt.Sprintf("User %s already exists. Skipping user creation.", s.config.User))
	} else {
		commands := []string{
			fmt.Sprintf("adduser --gecos '' --disabled-password %s", s.config.User),
			fmt.Sprintf("echo '%s:%s' | chpasswd", s.config.User, s.config.Passwd),
		}

		err := s.client.RunCommandWithProgress(
			ctx,
			fmt.Sprintf("Creating user %s...", s.config.User),
			fmt.Sprintf("User %s created successfully.", s.config.User),
			commands,
		)
		if err != nil {
			return err
		}
	}

	addToDockerCmd := fmt.Sprintf("usermod -aG docker %s", s.config.User)
	return s.client.RunCommandWithProgress(
		ctx,
		fmt.Sprintf("Adding user %s to Docker group...", s.config.User),
		fmt.Sprintf("User %s added to Docker group successfully.", s.config.User),
		[]string{addToDockerCmd},
	)
}

func (s *Server) setupServerSSHKey(ctx context.Context) error {
	keyPath := s.config.SSHKey
	if strings.HasPrefix(keyPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		keyPath = filepath.Join(homeDir, keyPath[1:])
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read SSH key: %w", err)
	}

	privKey, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("parse user private key for server access: %w", err)
	}
	pubKeyString := string(ssh.MarshalAuthorizedKey(privKey.PublicKey()))

	user := s.config.User
	sshDir := fmt.Sprintf("/home/%s/.ssh", user)
	authKeysFile := filepath.Join(sshDir, "authorized_keys")

	commands := []string{
		fmt.Sprintf("mkdir -p %s", sshDir),
		fmt.Sprintf("echo '%s' | tee -a %s", pubKeyString, authKeysFile),
		fmt.Sprintf("chown -R %s:%s %s", user, user, sshDir),
		fmt.Sprintf("chmod 700 %s", sshDir),
		fmt.Sprintf("chmod 600 %s", authKeysFile),
	}

	return s.client.RunCommandWithProgress(
		ctx,
		"Configuring SSH access for the new user...",
		"SSH access configured successfully.",
		commands,
	)
}

func (s *Server) configureDockerHub(ctx context.Context, dockerUsername, dockerPassword string) error {
	commands := []string{
		fmt.Sprintf("docker login -u %s -p %s", dockerUsername, dockerPassword),
	}

	return s.client.RunCommandWithProgress(
		ctx,
		"Logging into Docker Hub...",
		"Logged into Docker Hub successfully.",
		commands,
	)
}
