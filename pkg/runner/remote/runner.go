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

// ErrNoClient is returned when attempting operations on a closed Runner.
var ErrNoClient = errors.New("ssh client is nil")

// Runner executes commands and transfers files on a remote host via SSH.
// Once closed, a Runner cannot be reused.
type Runner struct {
	client *ssh.Client // client is unexported as it's an implementation detail
}

// NewRunner creates a new Runner instance using the provided SSH client.
// It returns nil if the client is nil.
func NewRunner(client *ssh.Client) *Runner {
	if client == nil {
		return nil
	}
	return &Runner{client: client}
}

// Close releases all resources associated with the Runner.
// After Close, the Runner cannot be reused.
func (r *Runner) Close() error {
	if r.client == nil {
		return nil
	}
	err := r.client.Close()
	r.client = nil
	return err
}

// RunCommands executes multiple commands sequentially on the remote host.
// It stops at the first command that fails.
func (r *Runner) RunCommands(ctx context.Context, commands []string) error {
	if r.client == nil {
		return ErrNoClient
	}

	for _, cmd := range commands {
		output, err := r.RunCommand(ctx, cmd)
		if err != nil {
			return fmt.Errorf("executing command %q: %w", cmd, err)
		}

		// Close the output reader in the same iteration to prevent resource leaks
		_, err = io.Copy(io.Discard, output)
		closeErr := output.Close()

		if err != nil {
			return fmt.Errorf("reading output of %q: %w", cmd, err)
		}
		if closeErr != nil {
			return fmt.Errorf("closing output of %q: %w", cmd, closeErr)
		}
	}
	return nil
}

// RunCommand executes a single command with optional arguments on the remote host.
// The caller must close the returned ReadCloser when done.
func (r *Runner) RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error) {
	if r.client == nil {
		return nil, ErrNoClient
	}

	session, err := r.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	// Build the full command with properly escaped arguments
	fullCmd := command
	if len(args) > 0 {
		escapedArgs := make([]string, len(args))
		for i, arg := range args {
			escapedArgs[i] = escapeArg(arg)
		}
		fullCmd += " " + strings.Join(escapedArgs, " ")
	}

	// Set up command I/O
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("creating stderr pipe: %w", err)
	}

	if err := session.Start(fullCmd); err != nil {
		session.Close()
		return nil, fmt.Errorf("starting command: %w", err)
	}

	return &commandOutput{
		reader:  io.MultiReader(stdout, stderr),
		session: session,
		ctx:     ctx,
	}, nil
}

// Host returns the hostname of the remote server.
func (r *Runner) Host() string {
	if r.client == nil {
		return ""
	}
	addr := r.client.RemoteAddr().String()
	host, _, _ := strings.Cut(addr, ":")
	return host
}

// CopyFile copies a file from src on the local machine to dst on the remote host.
// The destination file will have permissions 0644.
func (r *Runner) CopyFile(ctx context.Context, src, dst string) error {
	if r.client == nil {
		return ErrNoClient
	}

	client, err := scp.NewClientBySSH(r.client)
	if err != nil {
		return fmt.Errorf("creating SCP client: %w", err)
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer f.Close()

	if err := client.CopyFile(ctx, f, dst, "0644"); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

// commandOutput combines the stdout and stderr of a remote command
// and handles proper cleanup of the underlying SSH session.
type commandOutput struct {
	reader  io.Reader
	session *ssh.Session
	ctx     context.Context
}

func (c *commandOutput) Read(p []byte) (int, error) {
	// Check context cancellation before reading
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		return c.reader.Read(p)
	}
}

func (c *commandOutput) Close() error {
	// Send SIGTERM first for graceful shutdown
	_ = c.session.Signal(ssh.SIGTERM)

	var exitErr *ssh.ExitError
	err := c.session.Wait()
	if err != nil && !errors.As(err, &exitErr) {
		c.session.Close()
		return fmt.Errorf("waiting for command completion: %w", err)
	}

	return c.session.Close()
}

// escapeArg escapes a command-line argument for safe use in SSH commands.
func escapeArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}
