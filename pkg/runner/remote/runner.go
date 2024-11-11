package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

// Runner provides methods to run commands and copy files on a remote host.
type Runner struct {
	sshClient *ssh.Client
}

// NewRunner creates a new Runner instance.
func NewRunner(sshClient *ssh.Client) *Runner {
	return &Runner{
		sshClient: sshClient,
	}
}

// Close closes the SSH client.
func (c *Runner) Close() error {
	if c.sshClient == nil {
		return nil
	}

	err := c.sshClient.Close()
	c.sshClient = nil
	return err
}

// RunCommands runs multiple commands on the remote host.
func (c *Runner) RunCommands(ctx context.Context, commands []string) error {
	for _, command := range commands {
		rc, err := c.RunCommand(ctx, command)
		if err != nil {
			return fmt.Errorf("failed to run command '%s': %w", command, err)
		}

		_, readErr := io.ReadAll(rc)

		if readErr != nil {
			return fmt.Errorf("failed to read output of command '%s': %w", command, readErr)
		}
	}

	return nil
}

// RunCommand runs a command on the remote host.
func (c *Runner) RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %v", err)
	}

	fullCommand := command
	if len(args) > 0 {
		escapedArgs := make([]string, len(args))
		for i, arg := range args {
			escapedArgs[i] = sshEscapeArg(arg)
		}
		fullCommand += " " + strings.Join(escapedArgs, " ")
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := session.Start(fullCommand); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	outputReader := io.MultiReader(stdout, stderr)

	readCloser := &sessionReadCloser{
		Reader:  outputReader,
		session: session,
		ctx:     ctx,
	}

	return readCloser, nil
}

// sshEscapeArg properly escapes a command-line argument for SSH
func sshEscapeArg(arg string) string {
	return "'" + strings.Replace(arg, "'", "'\\''", -1) + "'"
}

// sessionReadCloser wraps an io.Reader and closes the SSH session when closed
type sessionReadCloser struct {
	io.Reader
	session *ssh.Session
	ctx     context.Context
}

// Close closes the SSH session and waits for the command to finish.
func (src *sessionReadCloser) Close() error {
	if err := src.session.Signal(ssh.SIGTERM); err != nil {
		src.session.Close()
	}

	if err := src.session.Wait(); err != nil {
		var exitMissingError *ssh.ExitMissingError
		if !errors.As(err, &exitMissingError) {
			return err
		}
	}

	return src.session.Close()
}

// CopyFile copies a file from the remote host to the local machine.
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
