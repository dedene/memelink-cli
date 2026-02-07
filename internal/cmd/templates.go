package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"github.com/dedene/memelink-cli/internal/actions"
	"github.com/dedene/memelink-cli/internal/api"
	"github.com/dedene/memelink-cli/internal/cache"
	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
	"github.com/dedene/memelink-cli/internal/preview"
	"github.com/dedene/memelink-cli/internal/tui"
	"github.com/dedene/memelink-cli/internal/ui"
)

// TemplatesCmd lists or views meme templates.
type TemplatesCmd struct {
	ID       string `arg:"" optional:"" help:"Template ID for detail view"`
	Filter   string `help:"Filter templates by name/keyword" name:"filter"`
	Animated bool   `help:"Show only animated-capable templates" name:"animated"`
	Refresh  bool   `help:"Force cache refresh" name:"refresh"`
}

// Run executes the templates command, dispatching to detail, interactive, or list view.
func (c *TemplatesCmd) Run(ctx context.Context, root *RootFlags) error {
	if c.ID != "" {
		return c.runDetail(ctx)
	}

	// TTY gate: interactive picker when stdout is terminal, not JSON, not --no-input, no --filter.
	if isatty.IsTerminal(os.Stdout.Fd()) && !outfmt.IsJSON(ctx) && !root.NoInput && c.Filter == "" {
		return c.runInteractive(ctx, root)
	}

	return c.runList(ctx)
}

// runDetail fetches a single template and prints its details.
func (c *TemplatesCmd) runDetail(ctx context.Context) error {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	tmpl, err := client.GetTemplate(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("getting template: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, tmpl)
	}

	fmt.Fprintf(os.Stdout, "ID:       %s\n", tmpl.ID)
	fmt.Fprintf(os.Stdout, "Name:     %s\n", tmpl.Name)
	fmt.Fprintf(os.Stdout, "Lines:    %d\n", tmpl.Lines)
	fmt.Fprintf(os.Stdout, "Overlays: %d\n", tmpl.Overlays)

	if len(tmpl.Styles) > 0 {
		fmt.Fprintf(os.Stdout, "Styles:   %s\n", strings.Join(tmpl.Styles, ", "))
	}

	if tmpl.Blank != "" {
		fmt.Fprintf(os.Stdout, "Blank:    %s\n", tmpl.Blank)
	}

	if tmpl.Example.URL != "" {
		fmt.Fprintf(os.Stdout, "Example:  %s\n", tmpl.Example.URL)
	}

	if len(tmpl.Keywords) > 0 {
		fmt.Fprintf(os.Stdout, "Keywords: %s\n", strings.Join(tmpl.Keywords, ", "))
	}

	if tmpl.Source != "" {
		fmt.Fprintf(os.Stdout, "Source:   %s\n", tmpl.Source)
	}

	return nil
}

// runInteractive launches the bubbletea fuzzy template picker with text input,
// then calls the generate API and prints the meme URL to stdout.
func (c *TemplatesCmd) runInteractive(ctx context.Context, root *RootFlags) error {
	templates, err := c.loadTemplates(ctx)
	if err != nil {
		return err
	}

	items := make([]list.Item, len(templates))
	for i, t := range templates {
		items[i] = tui.NewTemplateItem(t)
	}

	m := tui.NewPicker(items)

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr), tea.WithInputTTY())

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("interactive picker: %w", err)
	}

	picker, ok := result.(tui.Model)
	if !ok {
		return errors.New("unexpected picker result type")
	}

	if picker.Cancelled() || picker.Selected() == nil {
		return nil
	}

	// Generate meme with selected template + entered text.
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	cfg := config.FromContext(ctx)

	resp, err := client.Generate(ctx, api.GenerateRequest{
		TemplateID: picker.Selected().ID,
		Text:       picker.Texts(),
		Extension:  effectiveFormatFromConfig(cfg),
		Font:       effectiveFontFromConfig(cfg),
		Layout:     effectiveLayoutFromConfig(cfg),
		Redirect:   false,
	})
	if err != nil {
		return fmt.Errorf("generating meme: %w", err)
	}

	// Preview (config/default cascade only, no explicit flag on TemplatesCmd).
	if shouldPreview(nil, cfg, root) {
		_ = preview.Show(ctx, resp.URL, preview.Options{
			Writer: os.Stderr,
		})
	}

	fmt.Fprintln(os.Stdout, resp.URL)

	// Fire config-based auto actions (TUI flow has no explicit flags).
	if cfg != nil && cfg.AutoCopy != nil && *cfg.AutoCopy {
		if err := actions.CopyToClipboard(resp.URL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: clipboard: %v\n", err)
		}
	}

	if cfg != nil && cfg.AutoOpen != nil && *cfg.AutoOpen {
		if err := actions.OpenInBrowser(resp.URL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: browser: %v\n", err)
		}
	}

	return nil
}

// effectiveFormatFromConfig returns the config default format or "jpg".
func effectiveFormatFromConfig(cfg *config.Config) string {
	if cfg != nil && cfg.DefaultFormat != "" {
		return cfg.DefaultFormat
	}

	return "jpg"
}

// effectiveFontFromConfig returns the config default font or "" (API default).
func effectiveFontFromConfig(cfg *config.Config) string {
	if cfg != nil && cfg.DefaultFont != "" {
		return cfg.DefaultFont
	}

	return ""
}

// effectiveLayoutFromConfig returns the config default layout or "default".
func effectiveLayoutFromConfig(cfg *config.Config) string {
	if cfg != nil && cfg.DefaultLayout != "" {
		return cfg.DefaultLayout
	}

	return "default"
}

// runList fetches all templates and prints them as a table.
// Uses cached results when available and not --refresh.
func (c *TemplatesCmd) runList(ctx context.Context) error {
	templates, err := c.loadTemplates(ctx)
	if err != nil {
		return err
	}

	// Filter animated-capable if requested.
	if c.Animated {
		templates = filterAnimated(templates)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, templates)
	}

	// Build table rows.
	rows := make([][]string, 0, len(templates))
	for _, t := range templates {
		animated := ""
		if hasAnimated(t.Styles) {
			animated = "yes"
		}

		rows = append(rows, []string{t.ID, t.Name, fmt.Sprintf("%d", t.Lines), animated})
	}

	colorEnabled := false
	if u := ui.FromContext(ctx); u != nil {
		colorEnabled = u.Out().ColorEnabled()
	}

	fmt.Fprint(os.Stdout, ui.RenderTable(
		[]string{"ID", "Name", "Lines", "Animated"},
		rows,
		colorEnabled,
	))
	fmt.Fprintf(os.Stdout, "\n%d templates\n", len(templates))

	return nil
}

// loadTemplates fetches templates from cache or API. Shared by runList and runInteractive.
func (c *TemplatesCmd) loadTemplates(ctx context.Context) ([]api.Template, error) {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return nil, errors.New("api client not found in context")
	}

	var templates []api.Template

	// Try cache: only for unfiltered, non-refresh requests.
	if !c.Refresh && c.Filter == "" {
		if cached := c.loadCache(ctx); cached != nil {
			templates = cached
			slog.Debug("using cached templates", "count", len(templates))
		}
	}

	// Cache miss or bypass -- fetch from API.
	if templates == nil {
		var err error

		templates, err = client.ListTemplates(ctx, c.Filter)
		if err != nil {
			return nil, fmt.Errorf("listing templates: %w", err)
		}

		// Persist unfiltered results to cache (best-effort).
		if c.Filter == "" {
			c.saveCache(templates)
		}
	}

	return templates, nil
}

// loadCache attempts to load templates from disk cache.
// Returns nil on any error or cache miss.
func (c *TemplatesCmd) loadCache(ctx context.Context) []api.Template {
	cachePath, err := config.CachePath()
	if err != nil {
		return nil
	}

	ttl := 24 * time.Hour
	if cfg := config.FromContext(ctx); cfg != nil {
		ttl = cfg.CacheTTLDuration()
	}

	cached, err := cache.LoadTemplates(cachePath, ttl)
	if err != nil {
		slog.Debug("cache load error", "error", err)

		return nil
	}

	return cached
}

// saveCache persists templates to disk cache (best-effort).
func (c *TemplatesCmd) saveCache(templates []api.Template) {
	cachePath, err := config.CachePath()
	if err != nil {
		return
	}

	if err := cache.SaveTemplates(cachePath, templates); err != nil {
		slog.Debug("cache save error", "error", err)
	}
}

// hasAnimated checks if "animated" is present in a styles slice.
func hasAnimated(styles []string) bool {
	for _, s := range styles {
		if s == "animated" {
			return true
		}
	}

	return false
}

// filterAnimated returns only templates that have "animated" in their Styles.
func filterAnimated(templates []api.Template) []api.Template {
	result := make([]api.Template, 0, len(templates))
	for _, t := range templates {
		if hasAnimated(t.Styles) {
			result = append(result, t)
		}
	}

	return result
}
