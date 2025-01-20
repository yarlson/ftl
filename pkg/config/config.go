package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Project      Project      `yaml:"project" validate:"required"`
	Server       Server       `yaml:"server" validate:"required"`
	Services     []Service    `yaml:"services" validate:"required,dive"`
	Dependencies []Dependency `yaml:"dependencies" validate:"dive"`
	Volumes      []string     `yaml:"volumes" validate:"dive"`
}

type Project struct {
	Name   string `yaml:"name" validate:"required"`
	Domain string `yaml:"domain" validate:"required,fqdn"`
	Email  string `yaml:"email" validate:"required,email"`
}

type Server struct {
	Host       string `yaml:"host" validate:"required,fqdn|ip"`
	Port       int    `yaml:"port" validate:"required,min=1,max=65535"`
	User       string `yaml:"user" validate:"required"`
	Passwd     string `yaml:"-"`
	SSHKey     string `yaml:"ssh_key" validate:"required,filepath"`
	RootSSHKey string `yaml:"-"`
}

type Service struct {
	Name         string `yaml:"name" validate:"required"`
	Image        string `yaml:"image"`
	ImageUpdated bool
	Port         int                 `yaml:"port" validate:"required,min=1,max=65535"`
	Path         string              `yaml:"path"`
	HealthCheck  *ServiceHealthCheck `yaml:"health_check"`
	Routes       []Route             `yaml:"routes" validate:"required,dive"`
	Volumes      []string            `yaml:"volumes" validate:"dive,volume_reference"`
	Command      string              `yaml:"command"`
	Entrypoint   []string            `yaml:"entrypoint"`
	Env          []string            `yaml:"env"`
	Forwards     []string            `yaml:"forwards"`
	Recreate     bool                `yaml:"recreate"`
	Hooks        *Hooks              `yaml:"hooks"`
	Container    *Container          `yaml:"container"`
	LocalPorts   []int               `yaml:"-"`
}

type ServiceHealthCheck struct {
	Path     string        `yaml:"path"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Retries  int           `yaml:"retries"`
}

type Container struct {
	HealthCheck *ContainerHealthCheck `yaml:"health_check"`
	ULimits     []ULimit              `yaml:"ulimits"`
}

type ULimit struct {
	Name string `yaml:"name"`
	Hard int    `yaml:"hard"`
	Soft int    `yaml:"soft"`
}

type ContainerHealthCheck struct {
	Cmd          string `yaml:"cmd"`
	Interval     string `yaml:"interval"`
	Retries      int    `yaml:"retries"`
	Timeout      string `yaml:"timeout"`
	StartPeriod  string `yaml:"start_period"`
	StartTimeout string `yaml:"start_timeout"`
}

type Route struct {
	PathPrefix  string `yaml:"path" validate:"required"`
	StripPrefix bool   `yaml:"strip_prefix"`
}

type Dependency struct {
	Name      string     `yaml:"name" validate:"required"`
	Image     string     `yaml:"image" validate:"required"`
	Volumes   []string   `yaml:"volumes" validate:"dive,volume_reference"`
	Env       []string   `yaml:"env" validate:"dive"`
	Ports     []int      `yaml:"ports" validate:"dive,min=1,max=65535"`
	Container *Container `yaml:"container"`
}

// Hooks now supports either a simple remote command string
// or a map with local/remote commands.
type Hooks struct {
	Pre  *HookItem `yaml:"pre"`
	Post *HookItem `yaml:"post"`
}

// HookItem is used for either a single remote command (as a string),
// or a map of { remote, local } commands.
type HookItem struct {
	Remote string `yaml:"remote,omitempty"`
	Local  string `yaml:"local,omitempty"`
}

// UnmarshalYAML is a custom Unmarshaler to allow HookItem to be specified as a string or map.
func (h *HookItem) UnmarshalYAML(node *yaml.Node) error {
	switch node.Tag {
	case "!!str":
		// If user wrote: pre: "echo 'Pre-hook command'"
		// treat it as a remote command by default.
		if err := node.Decode(&h.Remote); err != nil {
			return err
		}
		return nil

	case "!!map":
		// If user wrote:
		// pre:
		//   remote: "echo 'Running remote'"
		//   local: "echo 'Running local'"
		type hookAlias HookItem
		var temp hookAlias
		if err := node.Decode(&temp); err != nil {
			return err
		}
		*h = HookItem(temp)
		return nil

	default:
		return fmt.Errorf("invalid hook format (must be string or map), got: %s", node.Tag)
	}
}

// getDefaultConfig retrieves a copy of the default config (if present),
// then applies the provided version to the Image.
func getDefaultConfig(baseName, version string) (*Dependency, bool) {
	baseName = strings.ToLower(baseName)
	dep, found := defaultConfigs[baseName]
	if !found {
		return nil, false
	}
	parts := strings.Split(dep.Image, ":")
	if len(parts) == 2 {
		dep.Image = parts[0] + ":" + version
	} else {
		dep.Image += ":" + version
	}
	return &dep, true
}

// UnmarshalYAML is a custom unmarshaler that handles both string-based
// dependencies (like "mysql:8") and map-based dependencies, plus expands env vars.
func (d *Dependency) UnmarshalYAML(node *yaml.Node) error {
	switch node.Tag {

	case "!!str":
		// If the node is just a string (e.g. "postgres:16"), parse it.
		value := node.Value
		parts := strings.SplitN(value, ":", 2)

		if len(parts) == 2 {
			// We have a base name + version
			base, version := parts[0], parts[1]
			if defaultDep, ok := getDefaultConfig(base, version); ok {
				// Expand env placeholders in the default config
				for i, envLine := range defaultDep.Env {
					expanded, err := expandWithEnvAndDefault(envLine)
					if err != nil {
						return fmt.Errorf(
							"failed expanding env in default config for %q: %w",
							base, err,
						)
					}
					defaultDep.Env[i] = expanded
				}
				*d = *defaultDep
			} else {
				// fallback for unknown base (e.g., "foobar:1.0")
				d.Name = base
				d.Image = value
			}
		} else {
			// Only a base name (e.g., "redis")
			base := parts[0]
			if defaultDep, ok := getDefaultConfig(base, "latest"); ok {
				// Expand env placeholders
				for i, envLine := range defaultDep.Env {
					expanded, err := expandWithEnvAndDefault(envLine)
					if err != nil {
						return fmt.Errorf(
							"failed expanding env in default config for %q: %w",
							base, err,
						)
					}
					defaultDep.Env[i] = expanded
				}
				*d = *defaultDep
			} else {
				// fallback for unknown base
				d.Name = base
				d.Image = base
			}
		}
		return nil

	case "!!map":
		// If the node is a map, decode into the struct in the usual way.
		type dependencyAlias Dependency
		var tmp dependencyAlias
		if err := node.Decode(&tmp); err != nil {
			return fmt.Errorf("failed to decode dependency map: %w", err)
		}
		// Expand placeholders in tmp.Env
		for i, envLine := range tmp.Env {
			expanded, err := expandWithEnvAndDefault(envLine)
			if err != nil {
				return fmt.Errorf(
					"failed to expand env for dependency %q: %w",
					tmp.Name, err,
				)
			}
			tmp.Env[i] = expanded
		}
		*d = Dependency(tmp)
		return nil

	default:
		// If there's some other type, return an error or handle as needed
		return fmt.Errorf("unsupported YAML type for Dependency: %s", node.Tag)
	}
}

type Volume struct {
	Name string `yaml:"name" validate:"required"`
	Path string `yaml:"path" validate:"required,unix_path"`
}

// expandWithEnvAndDefault expands environment variables within a single string.
// It handles `${VAR:-default}` and `${VAR:?error message}` syntax. If a required
// variable is missing, it returns an error. Otherwise, it returns the expanded
// string and a nil error.
func expandWithEnvAndDefault(input string) (string, error) {
	var expansionErr error

	expanded := os.Expand(input, func(key string) string {
		val, err := expandOneVar(key)
		if err != nil && expansionErr == nil {
			// capture the first error
			expansionErr = err
		}
		return val
	})

	// if expansionErr != nil, expanded might be partially filled, but we return the error
	return expanded, expansionErr
}

// expandOneVar handles a single ${...} expression inside os.Expand.
func expandOneVar(key string) (string, error) {
	// Check for ":-" = default fallback
	if strings.Contains(key, ":-") {
		parts := strings.SplitN(key, ":-", 2)
		envKey := parts[0]
		defaultVal := parts[1]
		if val, ok := os.LookupEnv(envKey); ok {
			return val, nil
		}
		return defaultVal, nil
	}

	// Check for ":?" = required variable
	if strings.Contains(key, ":?") {
		parts := strings.SplitN(key, ":?", 2)
		envKey := parts[0]
		errMsg := parts[1]
		if val, ok := os.LookupEnv(envKey); ok {
			return val, nil
		}
		// variable not set => return an error
		return "", fmt.Errorf("required environment variable %s not set: %s", envKey, errMsg)
	}

	// Otherwise, it's just ${VAR} with no colon
	if val, ok := os.LookupEnv(key); ok {
		return val, nil
	}
	// Not found in environment => return empty string
	return "", nil
}

func ParseConfig(data []byte) (*Config, error) {
	// Load any .env file from the current directory
	_ = godotenv.Load()

	// Process environment variables with default values
	expandedData, err := expandWithEnvAndDefault(string(data))
	if err != nil {
		return nil, fmt.Errorf("error expanding environment variables: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	// Process .env files for services if they exist
	for i := range config.Services {
		if config.Services[i].Path == "" {
			config.Services[i].Path = "./"
		}

		envPath := filepath.Join(config.Services[i].Path, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return nil, fmt.Errorf("failed to read .env file: %w", err)
			}
		}
	}

	validate := validator.New()

	// Register custom validations
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

	// Collect all named volumes from config.Services and config.Dependencies,
	// plus any that were explicitly listed in config.Volumes, deduplicating them.
	uniqueVolNames := make(map[string]struct{})

	// First, preserve any volumes already defined in config.Volumes
	for _, volName := range config.Volumes {
		uniqueVolNames[volName] = struct{}{}
	}

	// Check volumes in each service
	for _, svc := range config.Services {
		for _, volRef := range svc.Volumes {
			if volName := extractNamedVolume(volRef); volName != "" {
				uniqueVolNames[volName] = struct{}{}
			}
		}
	}

	// Check volumes in each dependency
	for _, dep := range config.Dependencies {
		for _, volRef := range dep.Volumes {
			if volName := extractNamedVolume(volRef); volName != "" {
				uniqueVolNames[volName] = struct{}{}
			}
		}
	}

	// Convert to a sorted slice
	finalVols := make([]string, 0, len(uniqueVolNames))
	for name := range uniqueVolNames {
		finalVols = append(finalVols, name)
	}
	sort.Strings(finalVols)
	config.Volumes = finalVols

	return &config, nil
}

// extractNamedVolume checks if volRef is in the form "NAME:/some/path"
// and if NAME starts with a letter. If so, it returns NAME; otherwise "".
func extractNamedVolume(volRef string) string {
	parts := strings.SplitN(volRef, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	namePart := parts[0]
	if len(namePart) == 0 {
		return ""
	}

	// Check if first character is a letter [a-zA-Z].
	first := namePart[0]
	if (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') {
		return namePart
	}

	// If it starts with '/', '.', or anything else, we treat it as a path, not a named volume
	return ""
}

func (s *Service) Hash() (string, error) {
	sortedService := s.sortServiceFields()
	bytes, err := json.Marshal(sortedService)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sorted service: %w", err)
	}

	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}

func (s *Service) sortServiceFields() map[string]interface{} {
	sorted := make(map[string]interface{})
	v := reflect.ValueOf(*s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(value)
			sorted[field.Name] = sortSlice(s)
		case reflect.Map:
			m := reflect.ValueOf(value)
			sorted[field.Name] = sortMap(m)
		default:
			sorted[field.Name] = value
		}
	}

	return sorted
}

func sortSlice(s reflect.Value) []interface{} {
	sorted := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		sorted[i] = s.Index(i).Interface()
	}
	sort.Slice(sorted, func(i, j int) bool {
		return fmt.Sprintf("%v", sorted[i]) < fmt.Sprintf("%v", sorted[j])
	})
	return sorted
}

func sortMap(m reflect.Value) map[string]interface{} {
	sorted := make(map[string]interface{})
	for _, key := range m.MapKeys() {
		sorted[key.String()] = m.MapIndex(key).Interface()
	}
	return sorted
}
