package dockersync

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yarlson/ftl/pkg/executor/ssh"
)

const (
	testImage = "golang:1.21-alpine"
	sshPort   = "22/tcp"
)

type testContainer struct {
	container testcontainers.Container
	sshPort   nat.Port
}

func setupTestContainer(t *testing.T) (*testContainer, error) {
	ctx := context.Background()

	// Build the test container image
	buildCtx, err := createBuildContext(t)
	require.NoError(t, err)
	defer os.RemoveAll(buildCtx)

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
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(sshPort))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	return &testContainer{
		container: container,
		sshPort:   mappedPort,
	}, nil
}

func createBuildContext(t *testing.T) (string, error) {
	dir, err := os.MkdirTemp("", "dockersync-test")
	if err != nil {
		return "", err
	}

	// Copy Dockerfile
	dockerfile := filepath.Join(dir, "Dockerfile")
	if err := copyFile("testdata/Dockerfile", dockerfile); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	// Copy entrypoint script
	entrypoint := filepath.Join(dir, "entrypoint.sh")
	if err := copyFile("testdata/entrypoint.sh", entrypoint); err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	return dir, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func setupTestImage(t *testing.T, dockerClient *client.Client) error {
	ctx := context.Background()

	// Pull test image
	reader, err := dockerClient.ImagePull(ctx, testImage, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)

	return nil
}

func TestImageSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test container
	t.Log("Setting up test container...")
	tc, err := setupTestContainer(t)
	require.NoError(t, err)
	defer func() { _ = tc.container.Terminate(context.Background()) }()

	// Create SSH client
	t.Log("Creating SSH client...")
	sshClient, err := ssh.ConnectWithUserPassword("127.0.0.1", tc.sshPort.Port(), "root", "testpassword")
	require.NoError(t, err)
	defer sshClient.Close()

	// Create Docker client
	t.Log("Creating Docker client...")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)
	defer dockerClient.Close()

	// Set up test image
	t.Log("Setting up test image...")
	err = setupTestImage(t, dockerClient)
	require.NoError(t, err)

	// Create temporary directories for test
	t.Log("Creating temporary directories...")
	localStore, err := os.MkdirTemp("", "dockersync-local")
	require.NoError(t, err)
	defer os.RemoveAll(localStore)

	remoteStore := "/tmp/dockersync-remote"

	// Initialize ImageSync
	cfg := Config{
		ImageName:   testImage,
		LocalStore:  localStore,
		RemoteStore: remoteStore,
		MaxParallel: 4,
	}

	sync := NewImageSync(cfg, sshClient)

	// Run sync
	t.Log("Running sync...")
	ctx := context.Background()
	err = sync.Sync(ctx)
	require.NoError(t, err)

	// Verify image exists on remote
	t.Log("Verifying image exists on remote...")
	output, err := sshClient.RunCommandOutput("docker images --format '{{.Repository}}:{{.Tag}}'")
	require.NoError(t, err)
	require.Contains(t, output, testImage)

	// Test image comparison
	t.Log("Comparing images...")
	needsSync, err := sync.compareImages(ctx)
	require.NoError(t, err)
	require.False(t, needsSync, "Images should be identical after sync")

	// Test re-sync with no changes
	t.Log("Re-syncing...")
	err = sync.Sync(ctx)
	require.NoError(t, err)
}
