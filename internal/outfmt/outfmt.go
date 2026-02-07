// Package outfmt provides context-based output mode selection (JSON vs human).
package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Mode controls output formatting.
type Mode struct {
	JSON bool
}

type ctxKey struct{}

// WithMode stores the output mode in the context.
func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, ctxKey{}, mode)
}

// IsJSON returns true if the context has JSON output mode enabled.
func IsJSON(ctx context.Context) bool {
	if v := ctx.Value(ctxKey{}); v != nil {
		if m, ok := v.(Mode); ok {
			return m.JSON
		}
	}

	return false
}

// WriteJSON writes v as pretty-printed JSON to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}
