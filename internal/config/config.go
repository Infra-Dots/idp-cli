package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = ".idp/config.yaml"

// Profile holds connection settings for one InfraDots environment.
type Profile struct {
	Host       string `yaml:"host"`
	Token      string `yaml:"token"`
	WebURL     string `yaml:"web_url,omitempty"`
	DefaultOrg string `yaml:"default_org,omitempty"`
}

// Config is the top-level structure of ~/.idp/config.yaml.
type Config struct {
	DefaultProfile string              `yaml:"default_profile"`
	Profiles       map[string]*Profile `yaml:"profiles"`
}

// ConfigPath returns the absolute path to the config file.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfigFile
	}
	return filepath.Join(home, defaultConfigFile)
}

// Load reads the config file. Returns an empty config if the file doesn't exist.
func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Profiles: map[string]*Profile{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	return &cfg, nil
}

// Save writes the config back to disk, creating parent directories as needed.
func Save(cfg *Config) error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ActiveProfile returns the profile to use, resolving the selection order:
// 1. profileName argument (from --profile flag)
// 2. INFRADOTS_PROFILE env var
// 3. cfg.DefaultProfile
// 4. "default"
func (cfg *Config) ActiveProfile(profileName string) *Profile {
	name := profileName
	if name == "" {
		name = os.Getenv("INFRADOTS_PROFILE")
	}
	if name == "" {
		name = cfg.DefaultProfile
	}
	if name == "" {
		name = "default"
	}
	p, ok := cfg.Profiles[name]
	if !ok {
		return &Profile{}
	}
	return p
}

// SetProfile upserts a profile in the config.
func (cfg *Config) SetProfile(name string, p *Profile) {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]*Profile{}
	}
	cfg.Profiles[name] = p
}

// RemoveToken clears the token from the named profile.
func (cfg *Config) RemoveToken(profileName string) {
	if p, ok := cfg.Profiles[profileName]; ok {
		p.Token = ""
	}
}
