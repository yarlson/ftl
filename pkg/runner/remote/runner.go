package remote

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
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

func (c *Runner) RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %v", err)
	}
	// Do not defer session.Close() here; we'll handle it in the ReadCloser.

	// Build the full command string
	fullCommand := command
	if len(args) > 0 {
		// Properly escape and join the arguments
		escapedArgs := make([]string, len(args))
		for i, arg := range args {
			escapedArgs[i] = sshEscapeArg(arg)
		}
		fullCommand += " " + strings.Join(escapedArgs, " ")
	}

	// Set up pipes for stdout and stderr
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

	// Start the command
	if err := session.Start(fullCommand); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Combine stdout and stderr into a single reader
	outputReader := io.MultiReader(stdout, stderr)

	// Create a ReadCloser that wraps the outputReader and closes the session when done
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

func (src *sessionReadCloser) Close() error {
	// Signal the session to terminate the command
	if err := src.session.Signal(ssh.SIGTERM); err != nil {
		// If signaling fails, forcibly close the session
		src.session.Close()
	}

	// Wait for the command to finish
	if err := src.session.Wait(); err != nil {
		if _, ok := err.(*ssh.ExitMissingError); !ok {
			// Ignore ExitMissingError which can occur if the session is closed prematurely
			return err
		}
	}

	// Close the session
	return src.session.Close()
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
