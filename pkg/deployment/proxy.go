package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/proxy"
)

func (d *Deployment) startProxy(ctx context.Context, project string, cfg *config.Config) error {
	hostname := d.runner.Host()

	// Prepare project folder
	projectPath, err := d.prepareProjectFolder(project)
	if err != nil {
		return fmt.Errorf("failed to prepare project folder: %w", err)
	}

	// Prepare nginx config
	spinner := d.sm.AddSpinner("config", fmt.Sprintf("[%s] Preparing Nginx configuration", hostname))
	configPath, err := d.prepareNginxConfig(cfg, projectPath)
	if err != nil {
		spinner.Error()
		return fmt.Errorf("failed to prepare nginx config: %w", err)
	}
	spinner.Complete()

	// Deploy proxy service
	spinner = d.sm.AddSpinner("proxy", fmt.Sprintf("[%s] Deploying proxy service", hostname))
	service := &config.Service{
		Name:  "proxy",
		Image: "yarlson/zero-nginx:latest",
		Port:  80,
		Volumes: []string{
			"certs:/etc/nginx/ssl",
			configPath + ":/etc/nginx/conf.d",
		},
		Env: []string{
			"DOMAIN=" + cfg.Project.Domain,
			"EMAIL=" + cfg.Project.Email,
		},
		Forwards: []string{
			"80:80",
			"443:443",
		},
		HealthCheck: &config.HealthCheck{
			Path:     "/",
			Interval: time.Second,
			Timeout:  time.Second,
			Retries:  30,
		},
		Recreate: true,
	}

	if err := d.deployService(project, service); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to deploy proxy service: %w", err)
	}
	spinner.Complete()

	// Reload nginx config
	spinner = d.sm.AddSpinner("nginx", fmt.Sprintf("[%s] Reloading Nginx configuration", hostname))
	if err := d.reloadNginxConfig(ctx); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to reload nginx config: %w", err)
	}
	spinner.Complete()

	// Deploy cert renewer
	spinner = d.sm.AddSpinner("certrenewer", fmt.Sprintf("[%s] Deploying certificate renewer", hostname))
	if err := d.deployCertRenewer(project, cfg); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to deploy certificate renewer: %w", err)
	}
	spinner.Complete()

	return nil
}

func (d *Deployment) prepareProjectFolder(project string) (string, error) {
	if err := d.makeProjectFolder(project); err != nil {
		return "", fmt.Errorf("failed to create project folder: %w", err)
	}

	return d.projectFolder(project)
}

func (d *Deployment) prepareNginxConfig(cfg *config.Config, projectPath string) (string, error) {
	nginxConfig, err := proxy.GenerateNginxConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to generate nginx config: %w", err)
	}

	nginxConfig = strings.TrimSpace(nginxConfig)

	configPath := filepath.Join(projectPath, "nginx")
	_, err = d.runCommand(context.Background(), "mkdir", "-p", configPath)
	if err != nil {
		return "", fmt.Errorf("failed to create nginx config directory: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "nginx-config-*.conf")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(nginxConfig); err != nil {
		return "", fmt.Errorf("failed to write nginx config to temporary file: %w", err)
	}

	return configPath, d.runner.CopyFile(context.Background(), tmpFile.Name(), filepath.Join(configPath, "default.conf"))
}

func (d *Deployment) reloadNginxConfig(ctx context.Context) error {
	_, err := d.runCommand(ctx, "docker", "exec", "proxy", "nginx", "-s", "reload")
	return err
}

func (d *Deployment) deployCertRenewer(project string, cfg *config.Config) error {
	service := &config.Service{
		Name:  "certrenewer",
		Image: "yarlson/zero-nginx:1.27-alpine3.19-zero0.2.0-0.2",
		Volumes: []string{
			"certs:/etc/nginx/ssl",
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		Env: []string{
			"DOMAIN=" + cfg.Project.Domain,
			"EMAIL=" + cfg.Project.Email,
			"PROXY_CONTAINER_NAME=proxy",
		},
		Entrypoint: []string{"/renew-certificates.sh"},
		Recreate:   true,
	}

	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to deploy certrenewer service: %w", err)
	}

	return nil
} 