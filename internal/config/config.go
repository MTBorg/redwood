package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Window struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type Config struct {
	Windows []Window `yaml:"windows"`
}

func configPath() string {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		cfg = filepath.Join(home, ".config")
	}
	return filepath.Join(cfg, "redwood", "config.yaml")
}

// Load reads the config file. If the file does not exist, an empty Config is returned.
func Load() (Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
