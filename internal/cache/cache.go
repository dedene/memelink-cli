// Package cache provides file-based template caching with TTL support.
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dedene/memelink-cli/internal/api"
)

// TemplateCache is the on-disk representation of cached templates.
type TemplateCache struct {
	Templates []api.Template `json:"templates"`
	FetchedAt time.Time      `json:"fetched_at"`
}

// LoadTemplates reads the cache file and returns templates if fresh.
// Returns (nil, nil) when: file missing, JSON corrupt, or TTL expired.
// Only returns a non-nil error for unexpected read failures.
func LoadTemplates(path string, ttl time.Duration) ([]api.Template, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is internal cache, not untrusted input
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("reading cache: %w", err)
	}

	var tc TemplateCache
	if err := json.Unmarshal(data, &tc); err != nil {
		// Corrupt cache -- treat as miss.
		return nil, nil //nolint:nilerr
	}

	if time.Since(tc.FetchedAt) > ttl {
		return nil, nil
	}

	return tc.Templates, nil
}

// SaveTemplates writes templates to the cache file atomically.
func SaveTemplates(path string, templates []api.Template) error {
	tc := TemplateCache{
		Templates: templates,
		FetchedAt: time.Now(),
	}

	data, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	data = append(data, '\n')

	return atomicWrite(path, data)
}

// atomicWrite writes data to path via temp-file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	tmpPath := tmp.Name()

	defer func() {
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	tmpPath = "" // prevent deferred cleanup

	return nil
}
