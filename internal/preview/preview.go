// Package preview renders inline terminal image previews.
package preview

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	// Register image format decoders.
	_ "image/jpeg"
	_ "image/png"

	termimg "github.com/blacktop/go-termimg"
	"golang.org/x/term"
)

// Options configures image preview rendering.
type Options struct {
	// Width in character cells. 0 = auto-detect from terminal.
	Width int
	// Writer receives rendered escape sequences. Typically os.Stderr.
	Writer io.Writer
}

// Show downloads an image from imageURL and renders it to opts.Writer.
// Returns nil on any error (download, decode, render) — never crashes.
func Show(ctx context.Context, imageURL string, opts Options) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	img, err := termimg.From(resp.Body)
	if err != nil {
		return nil
	}

	const (
		minPreviewWidth = 16
		maxPreviewWidth = 50
	)

	width := opts.Width
	if width <= 0 {
		w, _, sizeErr := term.GetSize(int(os.Stderr.Fd()))
		if sizeErr != nil || w <= 0 {
			width = 40
		} else {
			width = w / 3
		}
		width = max(minPreviewWidth, min(maxPreviewWidth, width))
	}

	rendered, err := img.Width(width).Scale(termimg.ScaleFit).Render()
	if err != nil {
		return nil
	}

	fmt.Fprintln(opts.Writer, rendered)

	return nil
}
