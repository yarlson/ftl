package network

import (
	"context"
)

// NetworkExists checks if a Docker network exists
func NetworkExists(ctx context.Context, network string) (bool, error) {
	// TODO: implement
	return false, nil
}

// CreateNetwork creates a new Docker network
func CreateNetwork(ctx context.Context, network string) error {
	// TODO: implement
	return nil
}
