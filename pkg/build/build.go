package build

import (
	"context"
	"fmt"
	"io"
)

type Executor interface {
	RunCommand(ctx context.Context, command string, args ...string) (io.Reader, error)
	RunCommands(ctx context.Context, commands []string) error
}

type Build struct {
	executor Executor
}

func NewBuild(executor Executor) *Build {
	return &Build{executor: executor}
}

func (b *Build) Build(ctx context.Context, image, path string) error {
	_, err := b.executor.RunCommand(ctx, "docker", "build", "-t", image, "--platform", "linux/amd64", path)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}

func (b *Build) Push(ctx context.Context, image string) error {
	_, err := b.executor.RunCommand(ctx, "docker", "push", image)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}
