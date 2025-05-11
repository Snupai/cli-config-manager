package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the dotman configuration
type Config struct {
	HomeDir    string
	DotmanDir  string
	ConfigsDir string
}

// NewWithoutDirectories creates a new Config without creating directories
func NewWithoutDirectories() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	dotmanDir := filepath.Join(homeDir, ".dotman")
	configsDir := filepath.Join(dotmanDir, "configs")

	return &Config{
		HomeDir:    homeDir,
		DotmanDir:  dotmanDir,
		ConfigsDir: configsDir,
	}, nil
}

// New creates a new Config and ensures all required directories exist
func New() (*Config, error) {
	cfg, err := NewWithoutDirectories()
	if err != nil {
		return nil, err
	}

	if err := cfg.EnsureDirectories(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// EnsureDirectories creates necessary directories if they don't exist
func (c *Config) EnsureDirectories() error {
	dirs := []string{c.DotmanDir, c.ConfigsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
