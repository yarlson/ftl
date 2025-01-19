package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestParseConfig_Success() {
	yamlPath := filepath.Join("sample", "ftl.yaml")
	yamlData, err := os.ReadFile(yamlPath)
	assert.NoError(suite.T(), err)

	config, err := ParseConfig(yamlData)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), "my-project", config.Project.Name)
	assert.Equal(suite.T(), "my-project.example.com", config.Project.Domain)
	assert.Equal(suite.T(), "my-project@example.com", config.Project.Email)
	assert.Len(suite.T(), config.Services, 1)
	assert.Equal(suite.T(), "my-app", config.Services[0].Name)
	assert.Equal(suite.T(), "my-app:latest", config.Services[0].Image)
	assert.Equal(suite.T(), 80, config.Services[0].Port)
	assert.Len(suite.T(), config.Services[0].Routes, 1)
	assert.Equal(suite.T(), "/", config.Services[0].Routes[0].PathPrefix)
	assert.False(suite.T(), config.Services[0].Routes[0].StripPrefix)
	assert.Len(suite.T(), config.Dependencies, 1)
	assert.Equal(suite.T(), "my-app-db", config.Dependencies[0].Name)
	assert.Equal(suite.T(), "my-app-db:latest", config.Dependencies[0].Image)
	assert.Len(suite.T(), config.Dependencies[0].Volumes, 1)
	assert.Equal(suite.T(), "my-app-db:/var/www/html/db", config.Dependencies[0].Volumes[0])
	assert.Len(suite.T(), config.Volumes, 1)
	assert.Equal(suite.T(), "my-app-db", config.Volumes[0])
}

func (suite *ConfigTestSuite) TestParseConfig_InvalidYAML() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"
services:
  - name: "web"
    image: "nginx:latest"
    routes:
      - path: "/"
        strip_prefix: true
        port: 80
  - this is invalid YAML
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "error parsing YAML")
}

func (suite *ConfigTestSuite) TestParseConfig_MissingRequiredFields() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
services:
  - name: "web"
    image: "nginx:latest"
    routes:
      - path: "/"
        strip_prefix: true
        port: 80
dependencies:
  - name: "db"
    image: "postgres:13"
    volumes:
      - "db_data:/var/lib/postgresql/data"
volumes:
  - db_data
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "validation error")
	assert.Contains(suite.T(), err.Error(), "Project.Email")
}

func (suite *ConfigTestSuite) TestParseConfig_InvalidEmail() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "invalid-email"
services:
  - name: "web"
    image: "nginx:latest"
    routes:
      - path: "/"
        strip_prefix: true
        port: 80
dependencies:
  - name: "db"
    image: "postgres:13"
    volumes:
      - "db_data:/var/lib/postgresql/data"
volumes:
  - db_data
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "validation error")
	assert.Contains(suite.T(), err.Error(), "Project.Email")
}

func (suite *ConfigTestSuite) TestParseConfig_InvalidVolumeReference() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
        strip_prefix: true
dependencies:
  - name: "db"
    image: "postgres:13"
    volumes:
      - "invalid_volume_reference"
volumes:
  - db_data
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "validation error")
	assert.Contains(suite.T(), err.Error(), "Config.Dependencies[0].Volumes[0]")
}

func (suite *ConfigTestSuite) TestParseConfig_WithHooks() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
    hooks:
      pre: "echo 'local pre-hook'"
      post: "echo 'local post-hook'"
`)

	config, err := ParseConfig(yamlData)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.NotNil(suite.T(), config.Services[0].Hooks)

	// Test Pre Hooks
	assert.NotNil(suite.T(), config.Services[0].Hooks.Pre)
	assert.Equal(suite.T(), "echo 'local pre-hook'", config.Services[0].Hooks.Pre)

	// Test Post Hooks
	assert.NotNil(suite.T(), config.Services[0].Hooks.Post)
	assert.Equal(suite.T(), "echo 'local post-hook'", config.Services[0].Hooks.Post)
}

func (suite *ConfigTestSuite) TestParseConfig_PartialHooks() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
    hooks:
      pre: "echo 'only local pre-hook'"
`)

	config, err := ParseConfig(yamlData)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.NotNil(suite.T(), config.Services[0].Hooks)

	// Test Pre Hooks
	assert.NotNil(suite.T(), config.Services[0].Hooks.Pre)
	assert.Equal(suite.T(), "echo 'only local pre-hook'", config.Services[0].Hooks.Pre)

	// Test Post Hooks
	assert.NotNil(suite.T(), config.Services[0].Hooks.Post)
	assert.Equal(suite.T(), "", config.Services[0].Hooks.Post)
}

func (suite *ConfigTestSuite) TestParseConfig_InvalidHookFormat() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"
servers:
  - host: "example.com"
    port: 22
    user: "user"
    ssh_key: "~/.ssh/id_rsa"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
    hooks:
      pre:
        invalid_field: true
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
}

func (suite *ConfigTestSuite) TestParseConfig_StringDependency() {
	yamlData := []byte(`
project:
  name: "string-dep-project"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
dependencies:
  - "postgres:16"
  - "redis"
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// We expect 2 dependencies
	assert.Len(suite.T(), config.Dependencies, 2)

	// Dependency #1: "postgres:16"
	assert.Equal(suite.T(), "postgres", config.Dependencies[0].Name)
	assert.Equal(suite.T(), "postgres:16", config.Dependencies[0].Image)

	// Dependency #2: "redis"
	// No colon, so we fall back to same name and image
	assert.Equal(suite.T(), "redis", config.Dependencies[1].Name)
	assert.Equal(suite.T(), "redis:latest", config.Dependencies[1].Image)
}

func (suite *ConfigTestSuite) TestParseConfig_MixedDependencies() {
	yamlData := []byte(`
project:
  name: "mixed-deps-project"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"
services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
dependencies:
  - "mysql:8"
  - name: "redis"
    image: "redis:6"
  - "elasticsearch"
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// We expect 3 dependencies
	assert.Len(suite.T(), config.Dependencies, 3)

	// Dependency #1: "mysql:8"
	assert.Equal(suite.T(), "mysql", config.Dependencies[0].Name)
	assert.Equal(suite.T(), "mysql:8", config.Dependencies[0].Image)

	// Dependency #2: normal map-based dependency
	assert.Equal(suite.T(), "redis", config.Dependencies[1].Name)
	assert.Equal(suite.T(), "redis:6", config.Dependencies[1].Image)

	// Dependency #3: "elasticsearch" (no colon)
	// We expect same name / image
	assert.Equal(suite.T(), "elasticsearch", config.Dependencies[2].Name)
	assert.Equal(suite.T(), "elasticsearch:latest", config.Dependencies[2].Image)
}

func (suite *ConfigTestSuite) TestParseConfig_VolumesExtraction() {
	yamlData := []byte(`
project:
  name: "test-project"
  domain: "example.com"
  email: "test@example.com"

server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"

services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    routes:
      - path: "/"
    volumes:
      - "my-vol:/app/data"
      - "/host/data:/container/data"  # not a named volume
  - name: "worker"
    image: "golang:latest"
    routes:
      - path: "/worker"
    port: 9000
    volumes:
      - "my-other-vol:/srv"

dependencies:
  - name: "some-db"
    image: "postgres:15"
    volumes:
      - "third-vol:/some/dep/path"
      - "123bad:/skip"
      - "logs:/var/log"

volumes:
  - "predefined-volume"
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	expected := []string{"logs", "my-other-vol", "my-vol", "predefined-volume", "third-vol"}
	assert.Equal(suite.T(), expected, config.Volumes)
}

func (suite *ConfigTestSuite) TestParseConfig_VolumesExtraction_NoNamedVolumes() {
	yamlData := []byte(`
project:
  name: "path-volumes-only"
  domain: "example.com"
  email: "test@example.com"

server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"

services:
  - name: "web"
    image: "nginx:latest"
    port: 80
    volumes:
      - "/absolute/path:/container/path"
      - "./relative:/opt/app"
dependencies:
  - name: "redis"
    image: "redis:latest"
    volumes:
      - "/redis/logs:/var/log/redis"
      - "./some-local-dir:/data"
volumes: []
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	assert.Empty(suite.T(), config.Volumes)
}

func (suite *ConfigTestSuite) TestParseConfig_VolumesExtraction_DefaultConfigs() {
	yamlData := []byte(`
project:
  name: "default-configs-only"
  domain: "example.com"
  email: "test@example.com"

server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"

services:
  - name: "web"
    image: "nginx:latest"
    routes:
      - path: "/"
    port: 80
dependencies:
  - "redis:6"
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	expected := []string{"redis_data"}
	assert.Equal(suite.T(), expected, config.Volumes)
}
