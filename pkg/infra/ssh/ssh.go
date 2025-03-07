package ssh

import (
	"context"

	"golang.org/x/crypto/ssh"
)

// NewSSHClientWithKey creates a new SSH client using key authentication
func NewSSHClientWithKey(ctx context.Context, host string, port int, user string, key []byte) (*ssh.Client, error) {
	// TODO: implement
	return nil, nil
}

// NewSSHClientWithPassword creates a new SSH client using password authentication
func NewSSHClientWithPassword(ctx context.Context, host string, port int, user, password string) (*ssh.Client, error) {
	// TODO: implement
	return nil, nil
}

// FindSSHKey reads and parses an SSH key from a file
func FindSSHKey(keyPath string) ([]byte, error) {
	// TODO: implement
	return nil, nil
}

// CopyFile copies a file from src to dst using SCP
func CopyFile(ctx context.Context, src, dst string) error {
	// TODO: implement
	return nil
}

// CreateSSHTunnel creates an SSH tunnel from local port to remote address
func CreateSSHTunnel(ctx context.Context, host string, port int, user, key string, localPort, remoteAddr string) error {
	// TODO: implement
	return nil
}
