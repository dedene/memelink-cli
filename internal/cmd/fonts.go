package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dedene/memelink-cli/internal/api"
	"github.com/dedene/memelink-cli/internal/outfmt"
	"github.com/dedene/memelink-cli/internal/ui"
)

// FontsCmd lists or views fonts.
type FontsCmd struct {
	ID string `arg:"" optional:"" help:"Font ID for detail view"`
}

// Run executes the fonts command, dispatching to detail or list view.
func (c *FontsCmd) Run(ctx context.Context) error {
	if c.ID != "" {
		return c.runDetail(ctx)
	}

	return c.runList(ctx)
}

// runDetail fetches a single font and prints its details.
func (c *FontsCmd) runDetail(ctx context.Context) error {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	font, err := client.GetFont(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("getting font: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, font)
	}

	fmt.Fprintf(os.Stdout, "ID:       %s\n", font.ID)

	alias := "-"
	if font.Alias != nil {
		alias = *font.Alias
	}

	fmt.Fprintf(os.Stdout, "Alias:    %s\n", alias)
	fmt.Fprintf(os.Stdout, "Filename: %s\n", font.Filename)

	return nil
}

// runList fetches all fonts and prints them as a table.
func (c *FontsCmd) runList(ctx context.Context) error {
	client := api.ClientFromContext(ctx)
	if client == nil {
		return errors.New("api client not found in context")
	}

	fonts, err := client.ListFonts(ctx)
	if err != nil {
		return fmt.Errorf("listing fonts: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, fonts)
	}

	rows := make([][]string, 0, len(fonts))
	for _, f := range fonts {
		alias := "-"
		if f.Alias != nil {
			alias = *f.Alias
		}

		rows = append(rows, []string{f.ID, alias, f.Filename})
	}

	colorEnabled := false
	if u := ui.FromContext(ctx); u != nil {
		colorEnabled = u.Out().ColorEnabled()
	}

	fmt.Fprint(os.Stdout, ui.RenderTable(
		[]string{"ID", "Alias", "Filename"},
		rows,
		colorEnabled,
	))
	fmt.Fprintf(os.Stdout, "\n%d fonts\n", len(fonts))

	return nil
}
