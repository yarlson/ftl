package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewSSHClientWithKey creates a new ssh.Client using a private key
func NewSSHClientWithKey(host string, port int, user string, key []byte) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", addr, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP connection: %v", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %v", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	return client, nil
}

// NewSSHClientWithPassword creates a new ssh.Client using a password
func NewSSHClientWithPassword(host string, port string, user string, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return client, nil
}

// sshKeyPath is used only for testing purposes
var sshKeyPath string

// FindSSHKey looks for an SSH key in the given path or in default locations
func FindSSHKey(keyPath string) ([]byte, error) {
	if keyPath != "" {
		if strings.HasPrefix(keyPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			keyPath = filepath.Join(home, keyPath[1:])
		}

		return os.ReadFile(keyPath)
	}

	sshDir, err := getSSHDir()
	if err != nil {
		return nil, err
	}

	keyNames := []string{"id_rsa", "id_ecdsa", "id_ed25519"}
	for _, name := range keyNames {
		path := filepath.Join(sshDir, name)
		key, err := os.ReadFile(path)
		if err == nil {
			return key, nil
		}
	}

	return nil, fmt.Errorf("no suitable SSH key found in %s", sshDir)
}

// FindKeyAndConnectWithUser finds an SSH key and establishes a connection
func FindKeyAndConnectWithUser(host string, port int, user, keyPath string) (*ssh.Client, []byte, error) {
	key, err := FindSSHKey(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find SSH key: %w", err)
	}

	client, err := NewSSHClientWithKey(host, port, user, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	return client, key, nil
}

// getSSHDir returns the SSH directory path
func getSSHDir() (string, error) {
	if sshKeyPath != "" {
		return sshKeyPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".ssh"), nil
}

// CreateSSHTunnel establishes an SSH tunnel from a local port to a remote address through an SSH server.
// It listens on localPort and forwards connections to remoteAddr via the SSH server at host:port.
// Authentication is done using the provided user and key (private key bytes).
func CreateSSHTunnel(host string, port int, user string, key []byte, localPort string, remoteAddr string) error {
	client, err := NewSSHClientWithKey(host, port, user, key)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %v", err)
	}

	localListener, err := net.Listen("tcp", "localhost:"+localPort)
	if err != nil {
		return fmt.Errorf("failed to listen on local port %s: %v", localPort, err)
	}
	defer localListener.Close()
	fmt.Printf("Listening on localhost:%s, forwarding to %s via %s:%d\n", localPort, remoteAddr, host, port)

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept local connection: %v\n", err)
			continue
		}

		remoteConn, err := client.Dial("tcp", remoteAddr)
		if err != nil {
			fmt.Printf("Failed to dial remote address %s: %v\n", remoteAddr, err)
			localConn.Close()
			continue
		}

		go func() {
			defer localConn.Close()
			defer remoteConn.Close()

			go func() {
				_, err := io.Copy(remoteConn, localConn)
				if err != nil {
					fmt.Printf("Error copying from local to remote: %v\n", err)
				}
			}()

			_, err := io.Copy(localConn, remoteConn)
			if err != nil {
				fmt.Printf("Error copying from remote to local: %v\n", err)
			}
		}()
	}
}
