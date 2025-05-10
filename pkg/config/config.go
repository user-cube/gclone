package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Profiles map[string]Profile `yaml:"profiles"`
}

// Profile represents a single profile configuration
type Profile struct {
	Name        string            `yaml:"name"`
	SSHHost     string            `yaml:"ssh_host"`
	URLPatterns []string          `yaml:"url_patterns"`
	GitConfigs  map[string]string `yaml:"git_configs"`
}

// DefaultConfigDir returns the default config directory path
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gclone"
	}
	return filepath.Join(home, ".gclone")
}

// DefaultConfigFile returns the default config file path
func DefaultConfigFile() string {
	return filepath.Join(DefaultConfigDir(), "config.yml")
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(configFile string) (*Config, error) {
	if configFile == "" {
		configFile = DefaultConfigFile()
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Profiles: make(map[string]Profile)}, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Initialize maps if they're nil
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	for name, profile := range config.Profiles {
		if profile.GitConfigs == nil {
			profile.GitConfigs = make(map[string]string)
			config.Profiles[name] = profile
		}
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *Config, configFile string) error {
	if configFile == "" {
		configFile = DefaultConfigFile()
	}

	// Ensure the directory exists
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Profiles: map[string]Profile{
			"personal": {
				Name:    "Personal",
				SSHHost: "github.com-personal",
				URLPatterns: []string{
					"github.com/your-personal-username",
					"github.com:your-personal-username",
				},
				GitConfigs: map[string]string{
					"user.name":  "Your Name",
					"user.email": "your.email@example.com",
				},
			},
			"work": {
				Name:    "Work",
				SSHHost: "github.com-work",
				URLPatterns: []string{
					"github.com/your-work-organization",
					"github.com:your-work-organization",
				},
				GitConfigs: map[string]string{
					"user.name":  "Your Work Name",
					"user.email": "your.work.email@example.com",
				},
			},
		},
	}
}
