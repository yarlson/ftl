package remote

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	ssh2 "github.com/yarlson/ftl/pkg/ssh"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Runner struct {
	sshClient *ssh.Client
}

func NewRunner(sshClient *ssh.Client) *Runner {
	return &Runner{
		sshClient: sshClient,
	}
}

func (c *Runner) Close() error {
	if c.sshClient == nil {
		return nil
	}

	err := c.sshClient.Close()
	c.sshClient = nil
	return err
}

func (c *Runner) RunCommand(ctx context.Context, command string, args ...string) (io.Reader, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %v", err)
	}
	defer session.Close()

	fullCommand := command
	if len(args) > 0 {
		var quotedArgs []string
		for _, arg := range args {
			quotedArgs = append(quotedArgs, fmt.Sprintf("%q", arg))
		}
		fullCommand += " " + strings.Join(quotedArgs, " ")
		fullCommand = strings.Replace(fullCommand, "\\n", "\n", -1)
	}

	pr, pw := io.Pipe()

	session.Stdout = pw
	session.Stderr = pw

	if err := session.Start(fullCommand); err != nil {
		_ = pw.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
		_ = pw.Close()
	}()

	var output bytes.Buffer
	outputChan := make(chan struct{})

	go func() {
		_, _ = io.Copy(&output, pr)
		close(outputChan)
	}()

	var commandErr error
	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		commandErr = ctx.Err()
	case err := <-done:
		if err != nil {
			commandErr = fmt.Errorf("command failed: %w", err)
		}
	}

	<-outputChan

	if commandErr != nil {
		return bytes.NewReader(output.Bytes()), commandErr
	}

	return bytes.NewReader(output.Bytes()), nil
}

func (c *Runner) CopyFile(ctx context.Context, src, dst string) error {
	client, err := scp.NewClientBySSH(c.sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SCP client: %w", err)
	}

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return client.CopyFile(ctx, file, dst, "0644")
}

func (c *Runner) RunCommands(ctx context.Context, commands []string) error {
	for _, command := range commands {
		if err := c.runSingleCommand(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (c *Runner) runSingleCommand(ctx context.Context, command string) error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("unable to create session: %w", err)
	}
	defer session.Close()

	pr, pw := io.Pipe()
	defer pr.Close()

	session.Stdout = pw
	session.Stderr = pw

	if err := session.Start(command); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
		_ = pw.Close()
	}()

	var output strings.Builder

	go func() {
		_, _ = io.Copy(&output, pr)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("%w\nOutput: %s", err, output.String())
		}
		return nil
	}
}

func (c *Runner) RunCommandWithOutput(command string) (string, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// sshKeyPath is used only for testing purposes
var sshKeyPath string

// FindSSHKey looks for an SSH key in the given path or in default locations
func FindSSHKey(keyPath string) ([]byte, error) {
	if keyPath != "" {
		if strings.HasPrefix(keyPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			keyPath = filepath.Join(home, keyPath[1:])
		}

		return os.ReadFile(keyPath)
	}

	sshDir, err := getSSHDir()
	if err != nil {
		return nil, err
	}

	keyNames := []string{"id_rsa", "id_ecdsa", "id_ed25519"}
	for _, name := range keyNames {
		path := filepath.Join(sshDir, name)
		key, err := os.ReadFile(path)
		if err == nil {
			return key, nil
		}
	}

	return nil, fmt.Errorf("no suitable SSH key found in %s", sshDir)
}

// FindKeyAndConnectWithUser finds an SSH key and establishes a connection
func FindKeyAndConnectWithUser(host string, port int, user, keyPath string) (*ssh.Client, []byte, error) {
	key, err := FindSSHKey(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find SSH key: %w", err)
	}

	client, err := ssh2.NewSSHClientWithKey(host, port, user, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	return client, key, nil
}

// getSSHDir returns the SSH directory path
func getSSHDir() (string, error) {
	if sshKeyPath != "" {
		return sshKeyPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".ssh"), nil
}

func (c *Runner) CreateTunnel(ctx context.Context, localPort, remotePort int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return fmt.Errorf("failed to start local listener: %w", err)
	}
	defer listener.Close()

	errChan := make(chan error, 1)

	go func() {
		for {
			local, err := listener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					errChan <- fmt.Errorf("failed to accept connection: %w", err)
				}
				return
			}

			go func(localConn net.Conn) {
				defer localConn.Close()

				remoteConn, err := c.sshClient.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", remotePort))
				if err != nil {
					fmt.Printf("Failed to connect to remote port: %v\n", err)
					return
				}
				defer remoteConn.Close()

				copyErrChan := make(chan error, 2)
				doneChan := make(chan bool, 2)

				go func() {
					_, err := io.Copy(localConn, remoteConn)
					copyErrChan <- err
					doneChan <- true
				}()

				go func() {
					_, err := io.Copy(remoteConn, localConn)
					copyErrChan <- err
					doneChan <- true
				}()

				select {
				case err := <-copyErrChan:
					if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
						fmt.Printf("Copy operation failed: %v\n", err)
					}
					<-doneChan
				case <-ctx.Done():
					return
				}
			}(local)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
