package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
)

func configTestCtx(t *testing.T, jsonMode bool) context.Context {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{}
	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: jsonMode})
	ctx = config.WithConfig(ctx, cfg)

	return ctx
}

func TestConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := &ConfigPathCmd{}
	output := captureStdout(t, func() {
		err := cmd.Run(context.Background())
		require.NoError(t, err)
	})

	assert.Contains(t, output, "memelink")
	assert.Contains(t, output, "config.json")
}

func TestConfigSetGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set
	setCmd := &ConfigSetCmd{Key: "default_format", Value: "png"}
	err := setCmd.Run(context.Background())
	require.NoError(t, err)

	// Load fresh from disk for get context
	cfgPath := filepath.Join(dir, "memelink", "config.json")
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = config.WithConfig(ctx, cfg)

	// Get
	getCmd := &ConfigGetCmd{Key: "default_format"}
	output := captureStdout(t, func() {
		err := getCmd.Run(ctx)
		require.NoError(t, err)
	})

	assert.Equal(t, "png\n", output)
}

func TestConfigList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set a value first
	setCmd := &ConfigSetCmd{Key: "default_format", Value: "webp"}
	require.NoError(t, setCmd.Run(context.Background()))

	// Load config for context
	cfgPath := filepath.Join(dir, "memelink", "config.json")
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: false})
	ctx = config.WithConfig(ctx, cfg)

	listCmd := &ConfigListCmd{}
	output := captureStdout(t, func() {
		err := listCmd.Run(ctx)
		require.NoError(t, err)
	})

	assert.Contains(t, output, "default_format = webp")
	assert.Contains(t, output, "safe = (unset)")
}

func TestConfigUnset(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set then unset
	setCmd := &ConfigSetCmd{Key: "default_format", Value: "png"}
	require.NoError(t, setCmd.Run(context.Background()))

	unsetCmd := &ConfigUnsetCmd{Key: "default_format"}
	require.NoError(t, unsetCmd.Run(context.Background()))

	// Load config for context
	cfgPath := filepath.Join(dir, "memelink", "config.json")
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = config.WithConfig(ctx, cfg)

	getCmd := &ConfigGetCmd{Key: "default_format"}
	output := captureStdout(t, func() {
		err := getCmd.Run(ctx)
		require.NoError(t, err)
	})

	assert.Equal(t, "(unset)\n", output)
}

func TestConfigSetInvalidKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setCmd := &ConfigSetCmd{Key: "invalid_key", Value: "foo"}
	err := setCmd.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigSetInvalidValue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	setCmd := &ConfigSetCmd{Key: "default_format", Value: "bmp"}
	err := setCmd.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}

func TestConfigListJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{DefaultFormat: "gif"}
	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})
	ctx = config.WithConfig(ctx, cfg)

	listCmd := &ConfigListCmd{}
	output := captureStdout(t, func() {
		err := listCmd.Run(ctx)
		require.NoError(t, err)
	})

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &parsed))
	assert.Equal(t, "gif", parsed["default_format"])
}

func TestConfigGetUnsetKey(t *testing.T) {
	ctx := configTestCtx(t, false)

	getCmd := &ConfigGetCmd{Key: "default_font"}
	output := captureStdout(t, func() {
		err := getCmd.Run(ctx)
		require.NoError(t, err)
	})

	assert.Equal(t, "(unset)\n", output)
}

func TestConfigUnsetInvalidKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	unsetCmd := &ConfigUnsetCmd{Key: "nope"}
	err := unsetCmd.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigFileCreated(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	setCmd := &ConfigSetCmd{Key: "default_font", Value: "impact"}
	require.NoError(t, setCmd.Run(context.Background()))

	cfgPath := filepath.Join(dir, "memelink", "config.json")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "impact", parsed["default_font"])
}
