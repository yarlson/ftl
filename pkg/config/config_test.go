package config

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
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
	assert.Equal(suite.T(), "echo 'local pre-hook'", config.Services[0].Hooks.Pre.Remote)

	// Test Post Hooks
	assert.NotNil(suite.T(), config.Services[0].Hooks.Post)
	assert.Equal(suite.T(), "echo 'local post-hook'", config.Services[0].Hooks.Post.Remote)
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
	assert.Equal(suite.T(), "echo 'only local pre-hook'", config.Services[0].Hooks.Pre.Remote)

	// Test Post Hooks
	assert.Nil(suite.T(), config.Services[0].Hooks.Post)
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
    routes:
      - path: "/"
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

func (suite *ConfigTestSuite) TestParseConfig_EnvExpansionInDefaults_Success() {
	// We'll set an env var that overrides the default "production-secret" for MySQL.
	// Then after the test, we'll unset it to avoid side effects in other tests.
	os.Setenv("MYSQL_ROOT_PASSWORD", "super-secret-password")
	defer os.Unsetenv("MYSQL_ROOT_PASSWORD")

	yamlData := []byte(`
project:
  name: "env-default-test"
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
  - "mysql:5.7"
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Len(suite.T(), config.Dependencies, 1)

	dep := config.Dependencies[0]
	// The name should come from the default config or be "mysql"
	assert.Equal(suite.T(), "mysql", dep.Name)
	// The image should reflect version override => "mysql:5.7"
	assert.Equal(suite.T(), "mysql:5.7", dep.Image)

	// Check env expansions
	// Our default config had "MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-production-secret}"
	// We set that env var => we expect "MYSQL_ROOT_PASSWORD=super-secret-password"
	require := assert.New(suite.T())
	require.Len(dep.Env, 1)
	require.Equal("MYSQL_ROOT_PASSWORD=super-secret-password", dep.Env[0])
}

func (suite *ConfigTestSuite) TestParseConfig_EnvExpansionInServices_Success() {
	// We'll set an env var that overrides a default in the service env.
	os.Setenv("MY_SERVICE_VAR", "overridden-value")
	defer os.Unsetenv("MY_SERVICE_VAR")

	yamlData := []byte(`
project:
  name: "service-env-test"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"

services:
  - name: "my-service"
    image: "my-service:latest"
    routes:
      - path: "/"
    port: 8080
    env:
      - "MY_VAR=${MY_SERVICE_VAR:-default-value}"
dependencies: []
`)

	config, err := ParseConfig(yamlData)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Len(suite.T(), config.Services, 1)

	svc := config.Services[0]
	require := assert.New(suite.T())
	require.Len(svc.Env, 1)
	// Should reflect the env var we set
	require.Equal("MY_VAR=overridden-value", svc.Env[0])
}

func (suite *ConfigTestSuite) TestParseConfig_EnvExpansion_RequiredVarMissing() {
	// We do NOT set the environment variable MY_REQUIRED_VAR, so we expect an error.
	yamlData := []byte(`
project:
  name: "required-var-test"
  domain: "example.com"
  email: "test@example.com"
server:
  host: "example.com"
  port: 22
  user: "user"
  ssh_key: "~/.ssh/id_rsa"

services:
  - name: "my-service"
    image: "my-service:latest"
    routes:
      - path: "/"
    port: 8080
    env:
      - MUST_BE_SET=${MY_REQUIRED_VAR:?must be set!}"
dependencies: []
`)

	config, err := ParseConfig(yamlData)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	// The error message should say "required environment variable MY_REQUIRED_VAR not set"
	assert.Contains(suite.T(), err.Error(), "required environment variable MY_REQUIRED_VAR not set")
	assert.Contains(suite.T(), err.Error(), "must be set!")
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid minimal config",
			yaml: `
project:
  name: test-project
  domain: example.com
  email: admin@example.com
server:
  host: server.example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
services:
  - name: web
    port: 8080
    routes:
      - path: /
`,
			want: &Config{
				Project: Project{
					Name:   "test-project",
					Domain: "example.com",
					Email:  "admin@example.com",
				},
				Server: Server{
					Host:   "server.example.com",
					Port:   22,
					User:   "deploy",
					SSHKey: "~/.ssh/id_rsa",
				},
				Services: []Service{
					{
						Name: "web",
						Port: 8080,
						Routes: []Route{
							{PathPrefix: "/"},
						},
					},
				},
			},
		},
		{
			name: "config with default host from project.domain",
			yaml: `
project:
  name: test-project
  domain: example.com
  email: admin@example.com
server:
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
services:
  - name: web
    port: 8080
    routes:
      - path: /
`,
			want: &Config{
				Project: Project{
					Name:   "test-project",
					Domain: "example.com",
					Email:  "admin@example.com",
				},
				Server: Server{
					Host:   "example.com",
					Port:   22,
					User:   "deploy",
					SSHKey: "~/.ssh/id_rsa",
				},
				Services: []Service{
					{
						Name: "web",
						Port: 8080,
						Routes: []Route{
							{PathPrefix: "/"},
						},
					},
				},
			},
		},
		{
			name: "config with environment variables",
			yaml: `
project:
  name: test-project
  domain: example.com
  email: admin@example.com
server:
  host: server.example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
services:
  - name: web
    port: 8080
    routes:
      - path: /
    env:
      - DB_HOST=${DB_HOST:-localhost}
      - DB_PORT=${DB_PORT:-5432}
      - API_KEY=${API_KEY:?API key is required}
`,
			envVars: map[string]string{
				"DB_HOST": "db.internal",
				"API_KEY": "secret123",
			},
			want: &Config{
				Project: Project{
					Name:   "test-project",
					Domain: "example.com",
					Email:  "admin@example.com",
				},
				Server: Server{
					Host:   "server.example.com",
					Port:   22,
					User:   "deploy",
					SSHKey: "~/.ssh/id_rsa",
				},
				Services: []Service{
					{
						Name: "web",
						Port: 8080,
						Routes: []Route{
							{PathPrefix: "/"},
						},
						Env: []string{
							"DB_HOST=db.internal",
							"DB_PORT=5432",
							"API_KEY=secret123",
						},
					},
				},
			},
		},
		{
			name: "config with dependencies",
			yaml: `
project:
  name: test-project
  domain: example.com
  email: admin@example.com
server:
  host: server.example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
services:
  - name: web
    port: 8080
    routes:
      - path: /
dependencies:
  - name: db
    image: postgres:13
    ports:
      - 5432
    volumes:
      - pg_data:/var/lib/postgresql/data
    env:
      - POSTGRES_PASSWORD=${DB_PASSWORD:-secret}
`,
			want: &Config{
				Project: Project{
					Name:   "test-project",
					Domain: "example.com",
					Email:  "admin@example.com",
				},
				Server: Server{
					Host:   "server.example.com",
					Port:   22,
					User:   "deploy",
					SSHKey: "~/.ssh/id_rsa",
				},
				Services: []Service{
					{
						Name: "web",
						Port: 8080,
						Routes: []Route{
							{PathPrefix: "/"},
						},
					},
				},
				Dependencies: []Dependency{
					{
						Name:  "db",
						Image: "postgres:13",
						Ports: []int{5432},
						Volumes: []string{
							"pg_data:/var/lib/postgresql/data",
						},
						Env: []string{
							"POSTGRES_PASSWORD=secret",
						},
					},
				},
			},
		},
		{
			name: "invalid config - missing required fields",
			yaml: `
project:
  name: test-project
services:
  - name: web
    routes:
      - path: /
`,
			wantErr: true,
		},
		{
			name: "invalid config - missing required env var",
			yaml: `
project:
  name: test-project
  domain: example.com
  email: admin@example.com
server:
  host: server.example.com
  port: 22
  user: deploy
  ssh_key: ~/.ssh/id_rsa
services:
  - name: web
    port: 8080
    routes:
      - path: /
    env:
      - REQUIRED_VAR=${REQUIRED_VAR:?This variable is required}
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, err := ParseConfig([]byte(tt.yaml))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.want.Server.User == "" {
				// If test expects empty user, replace with current user
				currentUser, err := user.Current()
				require.NoError(t, err)
				tt.want.Server.User = currentUser.Username
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServiceHash(t *testing.T) {
	tests := []struct {
		name     string
		service1 *Service
		service2 *Service
		wantSame bool
	}{
		{
			name: "identical services",
			service1: &Service{
				Name:  "web",
				Image: "nginx:latest",
				Port:  8080,
				Routes: []Route{
					{PathPrefix: "/"},
				},
			},
			service2: &Service{
				Name:  "web",
				Image: "nginx:latest",
				Port:  8080,
				Routes: []Route{
					{PathPrefix: "/"},
				},
			},
			wantSame: true,
		},
		{
			name: "different services",
			service1: &Service{
				Name:  "web",
				Image: "nginx:latest",
				Port:  8080,
			},
			service2: &Service{
				Name:  "web",
				Image: "nginx:1.19",
				Port:  8080,
			},
			wantSame: false,
		},
		{
			name: "same service with different ImageUpdated",
			service1: &Service{
				Name:         "web",
				Image:        "nginx:latest",
				ImageUpdated: true,
			},
			service2: &Service{
				Name:         "web",
				Image:        "nginx:latest",
				ImageUpdated: false,
			},
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err := tt.service1.Hash()
			require.NoError(t, err)

			hash2, err := tt.service2.Hash()
			require.NoError(t, err)

			if tt.wantSame {
				assert.Equal(t, hash1, hash2)
			} else {
				assert.NotEqual(t, hash1, hash2)
			}
		})
	}
}

func TestExpandWithEnvAndDefault(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		envVars map[string]string
		want    string
		wantErr bool
	}{
		{
			name:  "simple variable",
			input: "${VAR}",
			envVars: map[string]string{
				"VAR": "value",
			},
			want: "value",
		},
		{
			name:  "variable with default",
			input: "${VAR:-default}",
			want:  "default",
		},
		{
			name:  "variable with default but set",
			input: "${VAR:-default}",
			envVars: map[string]string{
				"VAR": "value",
			},
			want: "value",
		},
		{
			name:    "required variable not set",
			input:   "${VAR:?required}",
			wantErr: true,
		},
		{
			name:  "required variable set",
			input: "${VAR:?required}",
			envVars: map[string]string{
				"VAR": "value",
			},
			want: "value",
		},
		{
			name:  "multiple variables",
			input: "host=${HOST:-localhost} port=${PORT:-5432}",
			envVars: map[string]string{
				"HOST": "db.internal",
			},
			want: "host=db.internal port=5432",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, err := expandWithEnvAndDefault(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHookItemUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    *HookItem
		wantErr bool
	}{
		{
			name: "string hook",
			yaml: `"echo 'test'"`,
			want: &HookItem{
				Remote: "echo 'test'",
			},
		},
		{
			name: "map hook",
			yaml: `
remote: "echo 'remote'"
local: "echo 'local'"
`,
			want: &HookItem{
				Remote: "echo 'remote'",
				Local:  "echo 'local'",
			},
		},
		{
			name:    "invalid hook",
			yaml:    `[invalid]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got HookItem
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, &got)
		})
	}
}
