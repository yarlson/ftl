package proxy

import (
	"context"

	"github.com/yarlson/ftl/pkg/config"
)

// GenerateNginxConfig generates Nginx configuration from FTL config
func GenerateNginxConfig(cfg *config.Config) (string, error) {
	// TODO: implement
	return "", nil
}

// DeployProxy deploys or updates the Nginx reverse proxy container
func DeployProxy(ctx context.Context, project string, cfg *config.Config) error {
	// TODO: implement
	return nil
}
