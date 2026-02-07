package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/titanous/json5"
)

// Config holds user preferences.
type Config struct {
	DefaultFormat string `json:"default_format,omitempty"`
	DefaultFont   string `json:"default_font,omitempty"`
	DefaultLayout string `json:"default_layout,omitempty"`
	Safe          *bool  `json:"safe,omitempty"`
	AutoCopy      *bool  `json:"auto_copy,omitempty"`
	AutoOpen      *bool  `json:"auto_open,omitempty"`
	Preview       *bool  `json:"preview,omitempty"`
	CacheTTL      string `json:"cache_ttl,omitempty"`
}

// knownKey describes a config key and its optional validator.
type knownKey struct {
	validate func(string) error
}

var knownKeys = map[string]knownKey{
	"default_format": {validate: validateEnum("jpg", "png", "gif", "webp")},
	"default_font":   {validate: nil},
	"default_layout": {validate: validateEnum("default", "top")},
	"safe":           {validate: validateBool},
	"auto_copy":      {validate: validateBool},
	"auto_open":      {validate: validateBool},
	"preview":        {validate: validateBool},
	"cache_ttl":      {validate: validateDuration},
}

func validateEnum(allowed ...string) func(string) error {
	return func(val string) error {
		for _, a := range allowed {
			if val == a {
				return nil
			}
		}

		return fmt.Errorf("must be one of: %s", strings.Join(allowed, ", "))
	}
}

func validateBool(val string) error {
	if val != "true" && val != "false" {
		return fmt.Errorf("must be true or false")
	}

	return nil
}

func validateDuration(val string) error {
	_, err := time.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	return nil
}

// CacheTTLDuration parses CacheTTL as a time.Duration.
// Returns 24h on empty or invalid values.
func (cfg *Config) CacheTTLDuration() time.Duration {
	if cfg.CacheTTL == "" {
		return 24 * time.Hour
	}

	d, err := time.ParseDuration(cfg.CacheTTL)
	if err != nil {
		return 24 * time.Hour
	}

	return d
}

// Load reads config from the JSON5 file at path.
// Returns an empty Config if the file does not exist.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json5.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Save writes config as pretty-printed JSON atomically.
func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	data = append(data, '\n')

	return atomicWrite(path, data)
}

// atomicWrite writes data to path via temp-file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	tmpPath := tmp.Name()

	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()

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

// Get returns the string value for a config key and whether it is set.
func (cfg *Config) Get(key string) (string, bool) {
	switch key {
	case "default_format":
		return cfg.DefaultFormat, cfg.DefaultFormat != ""
	case "default_font":
		return cfg.DefaultFont, cfg.DefaultFont != ""
	case "default_layout":
		return cfg.DefaultLayout, cfg.DefaultLayout != ""
	case "safe":
		if cfg.Safe == nil {
			return "", false
		}

		return fmt.Sprintf("%t", *cfg.Safe), true
	case "auto_copy":
		if cfg.AutoCopy == nil {
			return "", false
		}

		return fmt.Sprintf("%t", *cfg.AutoCopy), true
	case "auto_open":
		if cfg.AutoOpen == nil {
			return "", false
		}

		return fmt.Sprintf("%t", *cfg.AutoOpen), true
	case "preview":
		if cfg.Preview == nil {
			return "", false
		}

		return fmt.Sprintf("%t", *cfg.Preview), true
	case "cache_ttl":
		return cfg.CacheTTL, cfg.CacheTTL != ""
	default:
		return "", false
	}
}

// Set sets a config key to a value after validation.
func (cfg *Config) Set(key, value string) error {
	kk, ok := knownKeys[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s (valid keys: %s)", key, strings.Join(KnownKeys(), ", "))
	}

	if kk.validate != nil {
		if err := kk.validate(value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	switch key {
	case "default_format":
		cfg.DefaultFormat = value
	case "default_font":
		cfg.DefaultFont = value
	case "default_layout":
		cfg.DefaultLayout = value
	case "safe":
		b := value == "true"
		cfg.Safe = &b
	case "auto_copy":
		b := value == "true"
		cfg.AutoCopy = &b
	case "auto_open":
		b := value == "true"
		cfg.AutoOpen = &b
	case "preview":
		b := value == "true"
		cfg.Preview = &b
	case "cache_ttl":
		cfg.CacheTTL = value
	}

	return nil
}

// Unset removes a config key (resets to zero/nil).
func (cfg *Config) Unset(key string) error {
	if _, ok := knownKeys[key]; !ok {
		return fmt.Errorf("unknown config key: %s (valid keys: %s)", key, strings.Join(KnownKeys(), ", "))
	}

	switch key {
	case "default_format":
		cfg.DefaultFormat = ""
	case "default_font":
		cfg.DefaultFont = ""
	case "default_layout":
		cfg.DefaultLayout = ""
	case "safe":
		cfg.Safe = nil
	case "auto_copy":
		cfg.AutoCopy = nil
	case "auto_open":
		cfg.AutoOpen = nil
	case "preview":
		cfg.Preview = nil
	case "cache_ttl":
		cfg.CacheTTL = ""
	}

	return nil
}

// KnownKeys returns a sorted list of valid config key names.
func KnownKeys() []string {
	keys := make([]string, 0, len(knownKeys))
	for k := range knownKeys {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// --- Context helpers ---

type ctxKey struct{}

// WithConfig stores a Config in the context.
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxKey{}, cfg)
}

// FromContext retrieves the Config from the context.
func FromContext(ctx context.Context) *Config {
	if v := ctx.Value(ctxKey{}); v != nil {
		if cfg, ok := v.(*Config); ok {
			return cfg
		}
	}

	return nil
}
