// Package config loads and parses mirdain.yaml.
package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds the HTTP server binding configuration.
type ServerConfig struct {
	Bind string `yaml:"bind"`
	Port int    `yaml:"port"`
}

// GitHubAppConfig holds GitHub credentials.
// In v1, only PAT is supported (TODO(#3): replace with App installation token).
type GitHubAppConfig struct {
	PAT string `yaml:"pat"`
}

// RepoConfig represents a single managed repository.
type RepoConfig struct {
	ID      string `yaml:"id"`
	Enabled bool   `yaml:"enabled"`
}

// Config is the top-level mirdain.yaml structure.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	GitHubApp GitHubAppConfig `yaml:"github_app"`
	Repos     []RepoConfig    `yaml:"repos"`
}

var envVarRe = regexp.MustCompile(`^\$\{([^}]+)\}$`)

// expandPAT expands a value of the form "${VAR}" to the environment variable
// named VAR. Only exact whole-value references are expanded to keep the
// contract explicit and avoid silently mutating unrelated fields.
func expandPAT(s string) string {
	m := envVarRe.FindStringSubmatch(s)
	if m == nil {
		return s
	}
	return os.Getenv(m[1])
}

// applyDefaults fills in built-in defaults for omitted fields.
func (c *Config) applyDefaults() {
	if c.Server.Bind == "" {
		c.Server.Bind = "127.0.0.1"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 7777
	}
}

// Load reads and parses the config file at path. It returns a descriptive
// error if the file is missing or malformed. Callers should log.Fatalf on
// the returned error; Load itself does not exit the process.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %q — copy mirdain.yaml.example to mirdain.yaml and fill in your values", path)
		}
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	cfg.GitHubApp.PAT = expandPAT(cfg.GitHubApp.PAT)
	cfg.applyDefaults()

	return &cfg, nil
}
