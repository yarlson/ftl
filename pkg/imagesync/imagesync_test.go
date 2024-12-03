package imagesync

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/yarlson/ftl/pkg/runner/remote"
	"github.com/yarlson/ftl/pkg/ssh"
	"github.com/yarlson/ftl/tests/dockercontainer"
)

const (
	testImage = "golang:1.21-alpine"
)

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
	tc, err := dockercontainer.NewContainer(t)
	require.NoError(t, err)
	defer func() { _ = tc.Container.Terminate(context.Background()) }()

	// Create SSH runner
	t.Log("Creating SSH runner...")
	sshClient, err := ssh.NewSSHClientWithPassword("127.0.0.1", tc.SshPort.Port(), "root", "testpassword")
	require.NoError(t, err)
	defer sshClient.Close()

	runner := remote.NewRunner(sshClient)

	// Create Docker runner
	t.Log("Creating Docker runner...")
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
		LocalStore:  localStore,
		RemoteStore: remoteStore,
		MaxParallel: 4,
	}

	sync := NewImageSync(cfg, runner)

	// Run sync
	t.Log("Running sync...")
	ctx := context.Background()
	_, err = sync.Sync(ctx, testImage)
	require.NoError(t, err)

	// Verify image exists on remote
	t.Log("Verifying image exists on remote...")
	outputReader, err := runner.RunCommand(ctx, "docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	require.NoError(t, err)
	defer outputReader.Close()

	output, err := io.ReadAll(outputReader)
	require.NoError(t, err)
	require.Contains(t, string(output), testImage)

	// Test image comparison
	t.Log("Comparing images...")
	needsSync, err := sync.CompareImages(ctx, testImage)
	require.NoError(t, err)
	require.False(t, needsSync, "Images should be identical after sync")

	// Test re-sync with no changes
	t.Log("Re-syncing...")
	_, err = sync.Sync(ctx, testImage)
	require.NoError(t, err)
}
