// Package config manages user preferences stored as JSON5/JSON files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the memelink config directory.
// Respects XDG_CONFIG_HOME; defaults to $HOME/.config/memelink.
func Dir() (string, error) {
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
func cacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "memelink"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "memelink"), nil
}

// Path returns the full path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.json"), nil
}

// CachePath returns the full path to the template cache file.
func CachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "templates.json"), nil
}
