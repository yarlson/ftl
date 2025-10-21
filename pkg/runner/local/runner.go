package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (e *Runner) RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	return io.NopCloser(bytes.NewReader(output)), nil
}

func (e *Runner) RunCommands(ctx context.Context, commands []string) error {
	operations := make([]func() error, len(commands))
	for i, cmdString := range commands {
		command, args := splitCommandAndArgs(cmdString)
		operations[i] = func() error {
			_, err := e.RunCommand(ctx, command, args...)
			return err
		}
	}

	return nil
}

func splitCommandAndArgs(cmdString string) (string, []string) {
	parts := strings.Fields(cmdString)
	if len(parts) == 0 {
		return "", nil
	}

	return parts[0], parts[1:]
}
