package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ssh2 "github.com/yarlson/ftl/pkg/ssh"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

func TestFindSSHKey(t *testing.T) {
	// Create temporary SSH keys in a temp directory
	tempDir := t.TempDir()
	sshKeyPath = tempDir

	keyContent := []byte("test-key")
	keyNames := []string{"id_rsa", "id_ecdsa", "id_ed25519"}

	// Write test keys
	for _, name := range keyNames {
		keyPath := filepath.Join(tempDir, name)
		err := os.WriteFile(keyPath, keyContent, 0600)
		assert.NoError(t, err)
	}

	// Override the home directory to point to tempDir
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	_ = os.Setenv("HOME", tempDir)

	// Test with no keyPath, should find id_rsa first
	key, err := FindSSHKey("")
	assert.NoError(t, err)
	assert.Equal(t, keyContent, key)

	// Test with specified keyPath
	specifiedKeyPath := filepath.Join(tempDir, "custom_key")
	err = os.WriteFile(specifiedKeyPath, keyContent, 0600)
	assert.NoError(t, err)

	key, err = FindSSHKey(specifiedKeyPath)
	assert.NoError(t, err)
	assert.Equal(t, keyContent, key)

	// Test when no keys are found
	_ = os.RemoveAll(tempDir)
	key, err = FindSSHKey("")
	assert.Error(t, err)
	assert.Nil(t, key)
}

func TestRunner_CreateTunnel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup
	tempDir := t.TempDir()
	privateKeyPath, publicKeyPath := filepath.Join(tempDir, "id_rsa"), filepath.Join(tempDir, "id_rsa.pub")
	require.NoError(t, generateSSHKeyPair(privateKeyPath, publicKeyPath))

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	require.NoError(t, err)

	imageTag := "ssh-nginx-test:latest"
	dockerfilePath := createDockerfile(t, strings.TrimSpace(string(publicKeyBytes)))
	buildImage(t, ctx, dockerfilePath, imageTag)

	// Start container
	container := startContainer(t, ctx, imageTag)
	defer func() { require.NoError(t, container.Terminate(ctx)) }()

	sshPort, err := container.MappedPort(ctx, "22")
	require.NoError(t, err)

	// Create SSH client
	key, err := os.ReadFile(privateKeyPath)
	require.NoError(t, err)

	client, err := ssh2.NewSSHClientWithKey("localhost", sshPort.Int(), "root", key)
	require.NoError(t, err)
	defer client.Close()

	tunnel := NewTunnel(client)

	// Test
	localPort, remotePort := 23451, 80
	tunnelCtx, tunnelCancel := context.WithCancel(ctx)
	defer tunnelCancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tunnel.CreateTunnel(tunnelCtx, localPort, remotePort)
	}()

	require.NoError(t, waitForPort(t, localPort, 5*time.Second))

	resp, err := makeHTTPRequest(localPort)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "<title>Welcome to nginx!</title>")

	// Cleanup
	tunnelCancel()
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Error("Tunnel didn't close within timeout")
	}
}

func waitForPort(t *testing.T, port int, timeout time.Duration) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("port %d not available after %s", port, timeout)
}

func startContainer(t *testing.T, ctx context.Context, imageTag string) testcontainers.Container {
	t.Helper()
	req := testcontainers.ContainerRequest{
		Image:        imageTag,
		ExposedPorts: []string{"22/tcp", "80/tcp"},
		WaitingFor:   wait.ForListeningPort("22/tcp").WithStartupTimeout(time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return container
}

func makeHTTPRequest(port int) (*http.Response, error) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	return httpClient.Get(fmt.Sprintf("http://localhost:%d", port))
}

func generateSSHKeyPair(privateKeyPath, publicKeyPath string) error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	privateKeyBytes, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: privateKeyBytes.Bytes,
	}

	privateKeyPEMBytes := pem.EncodeToMemory(privateKeyPEM)

	err = os.WriteFile(privateKeyPath, privateKeyPEMBytes, 0600)
	if err != nil {
		return err
	}

	publicKeySSH, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return err
	}

	return os.WriteFile(publicKeyPath, ssh.MarshalAuthorizedKey(publicKeySSH), 0644)
}

func createDockerfile(t *testing.T, publicKey string) string {
	content := `FROM ubuntu:20.04

RUN apt-get update && apt-get install -y openssh-server nginx

RUN mkdir /root/.ssh && \
    echo "%s" > /root/.ssh/authorized_keys && \
    chmod 700 /root/.ssh && \
    chmod 600 /root/.ssh/authorized_keys

RUN mkdir /run/sshd

EXPOSE 22 80

CMD service ssh start && nginx -g 'daemon off;'
`

	dockerfilePath := filepath.Join(t.TempDir(), "Dockerfile")
	err := os.WriteFile(dockerfilePath, []byte(fmt.Sprintf(content, publicKey)), 0644)
	require.NoError(t, err)
	return dockerfilePath
}

func buildImage(t *testing.T, ctx context.Context, dockerfilePath, imageTag string) {
	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)

	tar, err := archive.TarWithOptions(filepath.Dir(dockerfilePath), &archive.TarOptions{})
	require.NoError(t, err)

	opts := types.ImageBuildOptions{
		Dockerfile: filepath.Base(dockerfilePath),
		Tags:       []string{imageTag},
		Remove:     true,
	}

	resp, err := cli.ImageBuild(ctx, tar, opts)
	require.NoError(t, err)
	defer resp.Body.Close()

	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
}
