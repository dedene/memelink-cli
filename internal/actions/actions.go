// Package actions provides post-generation output actions: clipboard copy,
// browser open, and file download.
package actions

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"
)

// ErrClipboardUnsupported indicates the platform has no clipboard support.
var ErrClipboardUnsupported = errors.New("clipboard not supported on this platform")

// ErrHTTPStatus indicates the server returned a non-200 status code.
var ErrHTTPStatus = errors.New("unexpected HTTP status")

// ClipboardWrite is a function variable for clipboard writes (swappable in tests).
var ClipboardWrite = clipboard.WriteAll

// ClipboardUnsupported mirrors clipboard.Unsupported (swappable in tests).
var ClipboardUnsupported = clipboard.Unsupported

// BrowserOpen is a function variable for opening URLs (swappable in tests).
var BrowserOpen = browser.OpenURL

// CopyToClipboard copies text to the system clipboard.
// Returns a descriptive error if clipboard is unsupported on the platform.
func CopyToClipboard(text string) error {
	if ClipboardUnsupported {
		return ErrClipboardUnsupported
	}

	return ClipboardWrite(text)
}

// OpenInBrowser opens the given URL in the default browser.
func OpenInBrowser(rawURL string) error {
	return BrowserOpen(rawURL)
}

// DownloadFile downloads the resource at rawURL and saves it to destPath.
// Uses plain http.Get since meme URLs are public CDN resources.
func DownloadFile(rawURL, destPath string) error {
	resp, err := http.Get(rawURL) //nolint:gosec,noctx // public CDN URL, no auth needed
	if err != nil {
		return fmt.Errorf("downloading %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: %w: %d", rawURL, ErrHTTPStatus, resp.StatusCode)
	}

	f, err := os.Create(destPath) //nolint:gosec // destPath is user-provided output flag
	if err != nil {
		return fmt.Errorf("creating %s: %w", destPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing %s: %w", destPath, err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("syncing %s: %w", destPath, err)
	}

	return nil
}

// AutoFilename extracts a filename from a meme URL path.
// Falls back to "meme.jpg" if parsing fails or path is empty.
func AutoFilename(rawURL string) string {
	if rawURL == "" {
		return "meme.jpg"
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "meme.jpg"
	}

	base := path.Base(u.Path)
	if base == "" || base == "." || base == "/" {
		return "meme.jpg"
	}

	return base
}
