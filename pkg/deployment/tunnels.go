package deployment

import (
	"context"
	"fmt"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/tunnel"
)

func hasLocalHooks(cfg *config.Config) bool {
	for _, service := range cfg.Services {
		if service.Hooks != nil && service.Hooks.Pre != nil && service.Hooks.Pre.Local != "" {
			return true
		}

		if service.Hooks != nil && service.Hooks.Post != nil && service.Hooks.Post.Local != "" {
			return true
		}
	}

	return false
}

func (d *Deployment) startTunnels(ctx context.Context, cfg *config.Config) error {
	err := tunnel.StartTunnels(
		ctx,
		cfg.Server.Host, cfg.Server.Port,
		cfg.Server.User, cfg.Server.SSHKey,
		tunnel.CollectDependencyTunnels(cfg),
	)
	if err != nil {
		return fmt.Errorf("failed to establish tunnels: %w", err)
	}

	return nil
}
