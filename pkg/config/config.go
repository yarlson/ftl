package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Project  Project   `yaml:"project" validate:"required"`
	Servers  []Server  `yaml:"servers" validate:"required,dive"`
	Services []Service `yaml:"services" validate:"required,dive"`
	Storages []Storage `yaml:"storages" validate:"dive"`
	Volumes  []string  `yaml:"volumes" validate:"dive"`
}

type Project struct {
	Name   string `yaml:"name" validate:"required"`
	Domain string `yaml:"domain" validate:"required,fqdn"`
	Email  string `yaml:"email" validate:"required,email"`
}

type Server struct {
	Host   string `yaml:"host" validate:"required,fqdn|ip"`
	Port   int    `yaml:"port" validate:"required,min=1,max=65535"`
	User   string `yaml:"user" validate:"required"`
	SSHKey string `yaml:"ssh_key" validate:"required,filepath"`
}

type Service struct {
	Name        string       `yaml:"name" validate:"required"`
	Image       string       `yaml:"image" validate:"required"`
	Port        int          `yaml:"port" validate:"required,min=1,max=65535"`
	Path        string       `yaml:"path"`
	HealthCheck *HealthCheck `yaml:"health_check"`
	Routes      []Route      `yaml:"routes" validate:"required,dive"`
	Volumes     []string     `yaml:"volumes" validate:"dive,volume_reference"`

	Forwards []string

	EnvVars map[string]string
}

type EnvVar struct {
	Name  string
	Value string
}

type HealthCheck struct {
	Path     string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

type Route struct {
	PathPrefix  string `yaml:"path" validate:"required"`
	StripPrefix bool   `yaml:"strip_prefix"`
}

type Storage struct {
	Name    string            `yaml:"name" validate:"required"`
	Image   string            `yaml:"image" validate:"required"`
	Volumes []string          `yaml:"volumes" validate:"dive,volume_reference"`
	EnvVars map[string]string `yaml:"env" validate:"dive"`
}

type Volume struct {
	Name string `yaml:"name" validate:"required"`
	Path string `yaml:"path" validate:"required,unix_path"`
}

func ParseConfig(data []byte) (*Config, error) {
	expandedData := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	for service := range config.Services {
		if config.Services[service].Path == "" {
			config.Services[service].Path = "./"
		}
		envPath := filepath.Join(config.Services[service].Path, ".env")
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			continue
		}
		envMap, err := godotenv.Read(envPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}

		if config.Services[service].EnvVars == nil {
			config.Services[service].EnvVars = make(map[string]string)
		}
		for key, value := range envMap {
			config.Services[service].EnvVars[key] = value
		}
	}

	validate := validator.New()

	_ = validate.RegisterValidation("volume_reference", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		parts := strings.Split(value, ":")
		return len(parts) == 2 && parts[0] != "" && parts[1] != ""
	})

	_ = validate.RegisterValidation("unix_path", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return strings.HasPrefix(value, "/")
	})

	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("validation error: %v", err)
	}

	return &config, nil
}
