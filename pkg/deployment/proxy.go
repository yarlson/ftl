package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/proxy"
)

func (d *Deployment) startProxy(ctx context.Context, project string, cfg *config.Config) error {
	// Prepare project folder
	projectPath, err := d.prepareProjectFolder(project)
	if err != nil {
		return fmt.Errorf("failed to prepare project folder: %w", err)
	}

	// Prepare nginx config
	configPath, err := d.prepareNginxConfig(cfg, projectPath)
	if err != nil {
		return fmt.Errorf("failed to prepare nginx config: %w", err)
	}

	if err := d.deployZero(project, cfg); err != nil {
		return fmt.Errorf("failed to deploy Zero certificate manager: %w", err)
	}

	service := &config.Service{
		Name:  "proxy",
		Image: "nginx:alpine",
		Volumes: []string{
			"certs:/etc/nginx/certs:ro",
			configPath + ":/etc/nginx/conf.d:ro",
		},
		Forwards: []string{
			"443:443",
		},
		Container: &config.Container{
			HealthCheck: &config.ContainerHealthCheck{
				Cmd:      "curl -k https://localhost/",
				Interval: "10s",
				Retries:  3,
				Timeout:  "5s",
			},
		},
		Recreate: true,
	}

	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to deploy proxy service: %w", err)
	}

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

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.WriteString(nginxConfig); err != nil {
		return "", fmt.Errorf("failed to write nginx config to temporary file: %w", err)
	}

	return configPath, d.runner.CopyFile(context.Background(), tmpFile.Name(), filepath.Join(configPath, "default.conf"))
}

func (d *Deployment) deployZero(project string, cfg *config.Config) error {
	service := &config.Service{
		Name:  "zero",
		Image: "yarlson/zero:1",
		Volumes: []string{
			"certs:/certs",
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		Forwards: []string{
			"80:80",
		},
		CommandSlice: []string{
			"-d",
			cfg.Project.Domain,
			"-e",
			cfg.Project.Email,
			"-c",
			"/certs",
			"--hook",
			"nginx -s reload",
			"--hook-container",
			"proxy",
		},
		Recreate: true,
	}

	if err := d.deployService(project, service); err != nil {
		return fmt.Errorf("failed to deploy certrenewer service: %w", err)
	}

	return nil
}
