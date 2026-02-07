// Package config manages user preferences stored as JSON5/JSON files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigDir returns the memelink config directory.
// Respects XDG_CONFIG_HOME; defaults to $HOME/.config/memelink.
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "memelink"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, ".config", "memelink"), nil
}

// CacheDir returns the memelink cache directory.
// Respects XDG_CACHE_HOME; defaults to $HOME/.cache/memelink.
func CacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "memelink"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "memelink"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.json"), nil
}

// CachePath returns the full path to the template cache file.
func CachePath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "templates.json"), nil
}
