package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/config"
)

func TestLoadMissing(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "nonexistent", "config.json"))
	require.NoError(t, err)
	assert.Equal(t, &config.Config{}, cfg)
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	tr := true
	fa := false
	original := &config.Config{
		DefaultFormat: "png",
		DefaultFont:   "impact",
		DefaultLayout: "top",
		Safe:          &tr,
		AutoCopy:      &fa,
		AutoOpen:      nil,
		CacheTTL:      "12h",
	}

	require.NoError(t, config.Save(path, original))

	loaded, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, original.DefaultFormat, loaded.DefaultFormat)
	assert.Equal(t, original.DefaultFont, loaded.DefaultFont)
	assert.Equal(t, original.DefaultLayout, loaded.DefaultLayout)
	assert.Equal(t, original.CacheTTL, loaded.CacheTTL)
	require.NotNil(t, loaded.Safe)
	assert.True(t, *loaded.Safe)
	require.NotNil(t, loaded.AutoCopy)
	assert.False(t, *loaded.AutoCopy)
	assert.Nil(t, loaded.AutoOpen)
}

func TestLoadJSON5(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	json5Content := `{
		// User preferences
		"default_format": "png",
		"safe": true,  // trailing comma OK
	}`

	require.NoError(t, os.WriteFile(path, []byte(json5Content), 0o644))

	loaded, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "png", loaded.DefaultFormat)
	require.NotNil(t, loaded.Safe)
	assert.True(t, *loaded.Safe)
}

func TestGetSet(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{"default_format", "png"},
		{"default_font", "impact"},
		{"default_layout", "top"},
		{"safe", "true"},
		{"auto_copy", "false"},
		{"auto_open", "true"},
		{"cache_ttl", "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			cfg := &config.Config{}
			require.NoError(t, cfg.Set(tt.key, tt.value))

			got, ok := cfg.Get(tt.key)
			assert.True(t, ok)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestSetValidation(t *testing.T) {
	tests := []struct {
		key   string
		value string
		errRe string
	}{
		{"default_format", "bmp", "must be one of"},
		{"default_layout", "bottom", "must be one of"},
		{"safe", "yes", "must be true or false"},
		{"auto_copy", "1", "must be true or false"},
		{"cache_ttl", "forever", "invalid duration"},
		{"unknown_key", "foo", "unknown config key"},
	}

	for _, tt := range tests {
		t.Run(tt.key+"="+tt.value, func(t *testing.T) {
			cfg := &config.Config{}
			err := cfg.Set(tt.key, tt.value)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errRe)
		})
	}
}

func TestUnset(t *testing.T) {
	cfg := &config.Config{}
	require.NoError(t, cfg.Set("default_format", "png"))

	_, ok := cfg.Get("default_format")
	assert.True(t, ok)

	require.NoError(t, cfg.Unset("default_format"))

	_, ok = cfg.Get("default_format")
	assert.False(t, ok)
}

func TestUnsetUnknown(t *testing.T) {
	cfg := &config.Config{}
	err := cfg.Unset("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestBoolPointerDistinction(t *testing.T) {
	cfg := &config.Config{}

	// Unset: nil
	_, ok := cfg.Get("safe")
	assert.False(t, ok)
	assert.Nil(t, cfg.Safe)

	// Set false: non-nil false
	require.NoError(t, cfg.Set("safe", "false"))

	val, ok := cfg.Get("safe")
	assert.True(t, ok)
	assert.Equal(t, "false", val)
	require.NotNil(t, cfg.Safe)
	assert.False(t, *cfg.Safe)

	// Unset: back to nil
	require.NoError(t, cfg.Unset("safe"))
	assert.Nil(t, cfg.Safe)
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	path := filepath.Join(nested, "config.json")

	cfg := &config.Config{DefaultFormat: "gif"}
	require.NoError(t, config.Save(path, cfg))

	// Verify directory and file exist
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0)
}

func TestCacheTTLDuration(t *testing.T) {
	tests := []struct {
		name     string
		ttl      string
		expected time.Duration
	}{
		{"empty", "", 24 * time.Hour},
		{"valid", "1h", 1 * time.Hour},
		{"invalid", "forever", 24 * time.Hour},
		{"minutes", "30m", 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{CacheTTL: tt.ttl}
			assert.Equal(t, tt.expected, cfg.CacheTTLDuration())
		})
	}
}

func TestKnownKeys(t *testing.T) {
	keys := config.KnownKeys()
	assert.Len(t, keys, 8)

	// Verify sorted
	expected := []string{
		"auto_copy", "auto_open", "cache_ttl",
		"default_font", "default_format", "default_layout",
		"preview", "safe",
	}
	assert.Equal(t, expected, keys)
}

func TestConfigPaths(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	cfgPath, err := config.ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, cfgPath, "memelink")
	assert.Contains(t, cfgPath, "config.json")

	cachePath, err := config.CachePath()
	require.NoError(t, err)
	assert.Contains(t, cachePath, "memelink")
	assert.Contains(t, cachePath, "templates.json")
}

func TestConfigPathsDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	cfgPath, err := config.ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, cfgPath, ".config")
	assert.Contains(t, cfgPath, "memelink")

	cachePath, err := config.CachePath()
	require.NoError(t, err)
	assert.Contains(t, cachePath, ".cache")
	assert.Contains(t, cachePath, "memelink")
}

func TestWithConfig_FromContext(t *testing.T) {
	cfg := &config.Config{DefaultFormat: "webp"}
	ctx := config.WithConfig(context.Background(), cfg)

	got := config.FromContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "webp", got.DefaultFormat)
}

func TestFromContext_Nil(t *testing.T) {
	assert.Nil(t, config.FromContext(context.Background()))
}
