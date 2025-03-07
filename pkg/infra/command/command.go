package command

import (
	"context"
	"io"
)

// RunCommand executes a single command with arguments and returns its output
func RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error) {
	// TODO: implement
	return nil, nil
}

// RunCommands executes a slice of commands sequentially
func RunCommands(ctx context.Context, commands []string) error {
	// TODO: implement
	return nil
}
