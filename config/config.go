package config

import (
	"os"
	"path/filepath"
)

// Config represents the dotman configuration
type Config struct {
	HomeDir    string
	DotmanDir  string
	ConfigsDir string
}

// New creates a new Config instance
func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dotmanDir := filepath.Join(homeDir, ".dotman")
	configsDir := filepath.Join(dotmanDir, "configs")

	return &Config{
		HomeDir:    homeDir,
		DotmanDir:  dotmanDir,
		ConfigsDir: configsDir,
	}, nil
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
