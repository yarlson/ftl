package hooks

import (
	"context"

	"github.com/yarlson/ftl/pkg/config"
)

// ProcessPreHooks executes pre-deployment hooks
func ProcessPreHooks(ctx context.Context, service *config.Service) error {
	// TODO: implement
	return nil
}

// ProcessPostHooks executes post-deployment hooks
func ProcessPostHooks(ctx context.Context, service *config.Service, containerID string) error {
	// TODO: implement
	return nil
}
