package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
)

// ConfigCmd groups configuration subcommands.
type ConfigCmd struct {
	Path  ConfigPathCmd  `cmd:"" help:"Show config file path"`
	List  ConfigListCmd  `cmd:"" help:"List all config values"`
	Get   ConfigGetCmd   `cmd:"" help:"Get a config value"`
	Set   ConfigSetCmd   `cmd:"" help:"Set a config value"`
	Unset ConfigUnsetCmd `cmd:"" help:"Unset a config value"`
}

// ConfigPathCmd prints the config file path.
type ConfigPathCmd struct{}

// Run prints the config file path.
func (c *ConfigPathCmd) Run(_ context.Context) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, path)

	return nil
}

// ConfigListCmd lists all config values.
type ConfigListCmd struct{}

// Run lists all config keys with their values.
func (c *ConfigListCmd) Run(ctx context.Context) error {
	cfg := config.FromContext(ctx)
	if cfg == nil {
		cfg = &config.Config{}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, cfg)
	}

	for _, key := range config.KnownKeys() {
		val, ok := cfg.Get(key)
		if !ok {
			val = "(unset)"
		}

		fmt.Fprintf(os.Stdout, "%s = %s\n", key, val)
	}

	return nil
}

// ConfigGetCmd gets a single config value.
type ConfigGetCmd struct {
	Key string `arg:"" help:"Config key to get"`
}

// Run prints the value for the given key.
func (c *ConfigGetCmd) Run(ctx context.Context) error {
	cfg := config.FromContext(ctx)
	if cfg == nil {
		cfg = &config.Config{}
	}

	val, ok := cfg.Get(c.Key)
	if !ok {
		fmt.Fprintln(os.Stdout, "(unset)")

		return nil
	}

	fmt.Fprintln(os.Stdout, val)

	return nil
}

// ConfigSetCmd sets a config value.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key"`
	Value string `arg:"" help:"Config value"`
}

// Run sets a config key to a value, persisting to disk.
func (c *ConfigSetCmd) Run(_ context.Context) error {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	if err := cfg.Set(c.Key, c.Value); err != nil {
		return err
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Set %s = %s\n", c.Key, c.Value)

	return nil
}

// ConfigUnsetCmd removes a config value.
type ConfigUnsetCmd struct {
	Key string `arg:"" help:"Config key to unset"`
}

// Run unsets a config key, persisting to disk.
func (c *ConfigUnsetCmd) Run(_ context.Context) error {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	if err := cfg.Unset(c.Key); err != nil {
		return err
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Unset %s\n", c.Key)

	return nil
}
