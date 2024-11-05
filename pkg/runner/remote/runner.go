package remote

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
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
