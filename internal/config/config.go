package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	dirName    = "splitwise-cli"
	configFile = "config.json"
	authFile   = "auth.json"
)

// Config holds user preferences.
type Config struct {
	DefaultGroup    string `json:"default_group,omitempty"`
	DefaultCurrency string `json:"default_currency,omitempty"`
}

// Dir returns the config directory path (~/.config/splitwise-cli).
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, dirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", dirName)
}

// ConfigPath returns the full path to config.json.
func ConfigPath() string {
	return filepath.Join(Dir(), configFile)
}

// AuthPath returns the full path to auth.json.
func AuthPath() string {
	return filepath.Join(Dir(), authFile)
}

// Load reads config from disk. Returns zero-value Config if file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes config to disk.
func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0o600)
}

// Set updates a single config key.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	switch key {
	case "default_group":
		cfg.DefaultGroup = value
	case "default_currency":
		cfg.DefaultCurrency = value
	default:
		return &UnknownKeyError{Key: key}
	}
	return Save(cfg)
}

// UnknownKeyError is returned when an unrecognized config key is used.
type UnknownKeyError struct {
	Key string
}

func (e *UnknownKeyError) Error() string {
	return "unknown config key: " + e.Key
}
