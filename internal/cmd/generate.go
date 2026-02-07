package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/dedene/memelink-cli/internal/actions"
	"github.com/dedene/memelink-cli/internal/api"
	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
	"github.com/dedene/memelink-cli/internal/preview"
)

// validFormats lists accepted image formats.
var validFormats = map[string]bool{
	"jpg": true, "png": true, "gif": true, "webp": true,
}

// validLayouts lists accepted layout values.
var validLayouts = map[string]bool{
	"default": true, "top": true,
}

// GenerateCmd generates a meme. It is the default command when invoked
// with positional args (default:"withargs" in CLI struct).
type GenerateCmd struct {
	// Positional: first is template ID or auto-generate text; rest are text lines.
	Template string   `arg:"" optional:"" help:"Template ID (omit for auto-generate, 'custom' for custom background)"`
	Text     []string `arg:"" optional:"" help:"Text lines for the meme"`

	// Customization flags -- defaults empty; cascade fills from config/hardcoded.
	Format     string   `help:"Image format (jpg,png,gif,webp)" short:"f"`
	Font       string   `help:"Font ID or alias" name:"font"`
	TextColor  []string `help:"Text color per line (repeatable)" name:"text-color" sep:"none"`
	Layout     string   `help:"Text layout (default,top)" name:"layout"`
	Style      []string `help:"Style name or overlay URL (repeatable)" name:"style" sep:"none"`
	Width      int      `help:"Image width in pixels" name:"width"`
	Height     int      `help:"Image height in pixels" name:"height"`
	Center     string   `help:"Overlay center position (x,y)" name:"center"`
	Scale      string   `help:"Overlay scale ratio" name:"scale"`
	Safe       bool     `help:"Filter NSFW content" name:"safe"`
	Background string   `help:"Custom background image URL (use with 'custom' template)" name:"background"`

	// Output action flags.
	Copy       bool   `help:"Copy URL to clipboard" name:"copy" short:"c"`
	Open       bool   `help:"Open URL in browser" name:"open" short:"o"`
	Output     string `help:"Download image to file path" name:"output"`
	AutoOutput bool   `help:"Download image to CWD with auto-generated name" short:"O"`

	// Preview flag.
	Preview *bool `help:"Show inline image preview" name:"preview" negatable:""`
}

// shouldPreview determines if inline preview should be shown.
// Cascade: explicit flag > config preview > true (default ON for TTY).
// Always false when stderr is not a TTY or --no-input is set.
func shouldPreview(flag *bool, cfg *config.Config, root *RootFlags) bool {
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		return false
	}

	if root != nil && root.NoInput {
		return false
	}

	if flag != nil {
		return *flag
	}

	if cfg != nil && cfg.Preview != nil {
		return *cfg.Preview
	}

	return true
}

// Run executes the generate command, dispatching to one of three modes:
// auto-generate, template-based, or custom-background.
func (c *GenerateCmd) Run(ctx context.Context, root *RootFlags) error {
	if c.Template == "" && len(c.Text) == 0 {
		return errors.New("provide text or template ID; run 'memelink --help' for usage")
	}

	cfg := config.FromContext(ctx)

	// Validate effective format.
	format := c.effectiveFormat(cfg)
	if !validFormats[format] {
		return fmt.Errorf("invalid format %q: must be one of jpg, png, gif, webp", format)
	}

	// Validate effective layout.
	layout := c.effectiveLayout(cfg)
	if !validLayouts[layout] {
		return fmt.Errorf("invalid layout %q: must be one of default, top", layout)
	}

	// Auto-generate mode: single positional arg is the text.
	if c.Template != "" && len(c.Text) == 0 {
		return c.runAutomatic(ctx, cfg, root)
	}

	// Custom background mode.
	if c.Template == "custom" {
		return c.runCustom(ctx, cfg, root)
	}

	// Template-based mode.
	return c.runTemplate(ctx, cfg, root)
}

// effectiveFormat returns: explicit flag > config default > "jpg".
func (c *GenerateCmd) effectiveFormat(cfg *config.Config) string {
	if c.Format != "" {
		return c.Format
	}

	if cfg != nil && cfg.DefaultFormat != "" {
		return cfg.DefaultFormat
	}

	return "jpg"
}

// effectiveLayout returns: explicit flag > config default > "default".
func (c *GenerateCmd) effectiveLayout(cfg *config.Config) string {
	if c.Layout != "" {
		return c.Layout
	}

	if cfg != nil && cfg.DefaultLayout != "" {
		return cfg.DefaultLayout
	}

	return "default"
}

// effectiveFont returns: explicit flag > config default > "" (API default).
func (c *GenerateCmd) effectiveFont(cfg *config.Config) string {
	if c.Font != "" {
		return c.Font
	}

	if cfg != nil && cfg.DefaultFont != "" {
		return cfg.DefaultFont
	}

	return ""
}

// effectiveSafe returns: explicit --safe flag > config safe > false.
// Since bool default is false, config safe=true applies when flag not passed.
func (c *GenerateCmd) effectiveSafe(cfg *config.Config) bool {
	if c.Safe {
		return true
	}

	if cfg != nil && cfg.Safe != nil {
		return *cfg.Safe
	}

	return false
}

// effectiveCopy returns: explicit --copy flag > config auto_copy > false.
func (c *GenerateCmd) effectiveCopy(cfg *config.Config) bool {
	if c.Copy {
		return true
	}

	if cfg != nil && cfg.AutoCopy != nil {
		return *cfg.AutoCopy
	}

	return false
}

// effectiveOpen returns: explicit --open flag > config auto_open > false.
func (c *GenerateCmd) effectiveOpen(cfg *config.Config) bool {
	if c.Open {
		return true
	}

	if cfg != nil && cfg.AutoOpen != nil {
		return *cfg.AutoOpen
	}

	return false
}

// runActions fires post-generation actions (clipboard, browser, download).
// Errors are non-fatal warnings to stderr.
func (c *GenerateCmd) runActions(memeURL string, cfg *config.Config) {
	if c.effectiveCopy(cfg) {
		if err := actions.CopyToClipboard(memeURL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: clipboard: %v\n", err)
		}
	}

	if c.effectiveOpen(cfg) {
		if err := actions.OpenInBrowser(memeURL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: browser: %v\n", err)
		}
	}

	if c.Output != "" {
		if err := actions.DownloadFile(memeURL, c.Output); err != nil {
			fmt.Fprintf(os.Stderr, "warning: download: %v\n", err)
		}
	}

	if c.AutoOutput {
		if err := actions.DownloadFile(memeURL, actions.AutoFilename(memeURL)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: download: %v\n", err)
		}
	}
}

// runAutomatic calls POST /images/automatic with the provided text.
func (c *GenerateCmd) runAutomatic(ctx context.Context, cfg *config.Config, root *RootFlags) error {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	resp, err := client.GenerateAutomatic(ctx, api.AutomaticRequest{
		Text: c.Template,
		Safe: c.effectiveSafe(cfg),
	})
	if err != nil {
		return fmt.Errorf("generating meme: %w", err)
	}

	memeURL, err := api.AppendQueryParams(resp.URL, c.queryParams(cfg))
	if err != nil {
		return fmt.Errorf("appending query params: %w", err)
	}

	if shouldPreview(c.Preview, cfg, root) {
		_ = preview.Show(ctx, memeURL, preview.Options{
			Writer: os.Stderr,
		})
	}

	if outfmt.IsJSON(ctx) {
		if err := outfmt.WriteJSON(os.Stdout, map[string]any{
			"url":        memeURL,
			"generator":  resp.Generator,
			"confidence": resp.Confidence,
		}); err != nil {
			return err
		}

		c.runActions(memeURL, cfg)

		return nil
	}

	fmt.Fprintln(os.Stdout, memeURL)
	c.runActions(memeURL, cfg)

	return nil
}

// runTemplate calls POST /images for template-based meme generation.
func (c *GenerateCmd) runTemplate(ctx context.Context, cfg *config.Config, root *RootFlags) error {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	resp, err := client.Generate(ctx, api.GenerateRequest{
		TemplateID: c.Template,
		Text:       c.Text,
		Extension:  c.effectiveFormat(cfg),
		Font:       c.effectiveFont(cfg),
		Layout:     c.effectiveLayout(cfg),
		Style:      c.Style,
		Redirect:   false,
	})
	if err != nil {
		return fmt.Errorf("generating meme: %w", err)
	}

	return c.outputURL(ctx, resp.URL, cfg, root)
}

// runCustom calls POST /images/custom for custom-background meme generation.
func (c *GenerateCmd) runCustom(ctx context.Context, cfg *config.Config, root *RootFlags) error {
	if c.Background == "" {
		return errors.New("--background required when using 'custom' template")
	}

	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	// CustomRequest.Style is a single string; join repeatable flag values.
	style := strings.Join(c.Style, ",")

	resp, err := client.GenerateCustom(ctx, api.CustomRequest{
		Background: c.Background,
		Text:       c.Text,
		Extension:  c.effectiveFormat(cfg),
		Font:       c.effectiveFont(cfg),
		Layout:     c.effectiveLayout(cfg),
		Style:      style,
		Redirect:   false,
	})
	if err != nil {
		return fmt.Errorf("generating meme: %w", err)
	}

	return c.outputURL(ctx, resp.URL, cfg, root)
}

// outputURL appends query params, prints the meme URL, and fires actions.
func (c *GenerateCmd) outputURL(ctx context.Context, rawURL string, cfg *config.Config, root *RootFlags) error {
	memeURL, err := api.AppendQueryParams(rawURL, c.queryParams(cfg))
	if err != nil {
		return fmt.Errorf("appending query params: %w", err)
	}

	if shouldPreview(c.Preview, cfg, root) {
		_ = preview.Show(ctx, memeURL, preview.Options{
			Writer: os.Stderr,
		})
	}

	if outfmt.IsJSON(ctx) {
		if err := outfmt.WriteJSON(os.Stdout, map[string]any{
			"url": memeURL,
		}); err != nil {
			return err
		}

		c.runActions(memeURL, cfg)

		return nil
	}

	fmt.Fprintln(os.Stdout, memeURL)
	c.runActions(memeURL, cfg)

	return nil
}

// queryParams builds url.Values from presentation flags that are appended
// as query parameters to the returned meme URL (not sent in the POST body).
func (c *GenerateCmd) queryParams(cfg *config.Config) url.Values {
	v := url.Values{}
	if len(c.TextColor) > 0 {
		v.Set("color", strings.Join(c.TextColor, ","))
	}

	if c.Width > 0 {
		v.Set("width", strconv.Itoa(c.Width))
	}

	if c.Height > 0 {
		v.Set("height", strconv.Itoa(c.Height))
	}

	if c.Center != "" {
		v.Set("center", c.Center)
	}

	if c.Scale != "" {
		v.Set("scale", c.Scale)
	}

	if c.effectiveSafe(cfg) {
		v.Set("safe", "true")
	}

	return v
}
