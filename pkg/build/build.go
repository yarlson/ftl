package build

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Runner interface {
	RunCommand(ctx context.Context, command string, args ...string) (io.ReadCloser, error)
	RunCommands(ctx context.Context, commands []string) error
}

type Build struct {
	runner Runner
}

func NewBuild(runner Runner) *Build {
	return &Build{runner: runner}
}

func (b *Build) Build(ctx context.Context, image, path string) error {
	labelKey := "org.opencontainers.image.vendor"
	labelValue := "ftl"

	_, err := b.runner.RunCommand(ctx,
		"docker", "build",
		"-t", image,
		"--platform", "linux/amd64",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		path,
	)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	outputReader, err := b.runner.RunCommand(ctx,
		"docker", "images",
		"--filter", "dangling=true",
		"--filter", fmt.Sprintf("label=%s=%s", labelKey, labelValue),
		"--format", "{{.ID}}",
	)
	if err != nil {
		return fmt.Errorf("failed to list images for cleanup: %w", err)
	}
	defer outputReader.Close()

	outputBytes, err := io.ReadAll(outputReader)
	if err != nil {
		return fmt.Errorf("failed to read output of docker images: %w", err)
	}

	imageIDs := strings.Fields(string(outputBytes))
	if len(imageIDs) == 0 {
		return nil
	}

	args := append([]string{"rmi", "--force"}, imageIDs...)
	_, err = b.runner.RunCommand(ctx, "docker", args...)
	if err != nil {
		return fmt.Errorf("failed to remove images: %w", err)
	}

	return nil
}

func (b *Build) Push(ctx context.Context, image string) error {
	_, err := b.runner.RunCommand(ctx, "docker", "push", image)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}
