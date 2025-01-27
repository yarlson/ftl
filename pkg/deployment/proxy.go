package deployment

import (
	"context"
	"fmt"
	"github.com/yarlson/ftl/pkg/config"
	"github.com/yarlson/ftl/pkg/proxy"
	"os"
	"path/filepath"
	"strings"
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

	spinner = d.sm.AddSpinner("zero", fmt.Sprintf("[%s] Deploying Zero certificate manager", hostname))
	if err := d.deployZero(project, cfg); err != nil {
		spinner.Error()
		return fmt.Errorf("failed to deploy Zero certificate manager: %w", err)
	}
	spinner.Complete()

	spinner = d.sm.AddSpinner("proxy", fmt.Sprintf("[%s] Deploying proxy service", hostname))
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
		spinner.Error()
		return fmt.Errorf("failed to deploy proxy service: %w", err)
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
