package servercontainer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	sshPort = "22/tcp"
)

type Container struct {
	Container testcontainers.Container
	SshPort   nat.Port
}

func NewContainer(t *testing.T) (*Container, error) {
	ctx := context.Background()

	buildCtx, err := createBuildContext()
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(buildCtx)
	}()

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    buildCtx,
				Dockerfile: "Dockerfile",
			},
			ExposedPorts: []string{sshPort},
			Privileged:   true, // Required for Docker daemon
			WaitingFor: wait.ForAll(
				wait.ForListeningPort(sshPort),
			),
			Env: map[string]string{
				"DOCKER_TLS_CERTDIR": "", // Disable TLS for testing
			},
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start Container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, sshPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	return &Container{
		Container: container,
		SshPort:   mappedPort,
	}, nil
}

func createBuildContext() (string, error) {
	dir, err := os.MkdirTemp("", "dockersync-test")
	if err != nil {
		return "", err
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	packageDir := filepath.Dir(currentFile)

	dockerfile := filepath.Join(dir, "Dockerfile")
	if err := copyFile(filepath.Join(packageDir, "docker", "Dockerfile"), dockerfile); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)

	return err
}
