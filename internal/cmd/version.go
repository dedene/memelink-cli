package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/dedene/memelink-cli/internal/outfmt"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

// VersionString returns a human-readable version string.
func VersionString() string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}
	if strings.TrimSpace(commit) == "" && strings.TrimSpace(date) == "" {
		return v
	}
	if strings.TrimSpace(commit) == "" {
		return fmt.Sprintf("%s (%s)", v, strings.TrimSpace(date))
	}
	if strings.TrimSpace(date) == "" {
		return fmt.Sprintf("%s (%s)", v, strings.TrimSpace(commit))
	}
	return fmt.Sprintf("%s (%s %s)", v, strings.TrimSpace(commit), strings.TrimSpace(date))
}

// VersionCmd prints version information.
type VersionCmd struct{}

// Run executes the version command.
func (c *VersionCmd) Run(ctx context.Context) error {
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"version": strings.TrimSpace(version),
			"commit":  strings.TrimSpace(commit),
			"date":    strings.TrimSpace(date),
			"go":      runtime.Version(),
		})
	}

	fmt.Fprintf(os.Stdout, "memelink %s\n", VersionString())
	if c := strings.TrimSpace(commit); c != "" {
		fmt.Fprintf(os.Stdout, "  commit: %s\n", c)
	}
	if d := strings.TrimSpace(date); d != "" {
		fmt.Fprintf(os.Stdout, "  date:   %s\n", d)
	}
	fmt.Fprintf(os.Stdout, "  go:     %s\n", runtime.Version())
	return nil
}
