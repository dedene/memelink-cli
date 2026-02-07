package actions

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"jpg with path segments", "https://api.memegen.link/images/drake/a/b.jpg", "b.jpg"},
		{"jpg with query params", "https://api.memegen.link/images/drake/a/b.jpg?width=400", "b.jpg"},
		{"png file", "https://api.memegen.link/images/buzz/hello.png", "hello.png"},
		{"empty string", "", "meme.jpg"},
		{"invalid URL", "://bad", "meme.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, AutoFilename(tt.url))
		})
	}
}

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	body := []byte("fake-image-data-12345")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(body) //nolint:errcheck
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.jpg")
	err := DownloadFile(srv.URL+"/test.jpg", dest)
	require.NoError(t, err)

	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, body, got)
}

func TestDownloadFileHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.jpg")
	err := DownloadFile(srv.URL+"/missing.jpg", dest)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrHTTPStatus)
}

func TestCopyToClipboard(t *testing.T) {
	origWrite := ClipboardWrite
	origUnsupported := ClipboardUnsupported
	defer func() {
		ClipboardWrite = origWrite
		ClipboardUnsupported = origUnsupported
	}()

	ClipboardUnsupported = false

	var captured string
	ClipboardWrite = func(text string) error {
		captured = text
		return nil
	}

	err := CopyToClipboard("https://example.com/meme.jpg")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/meme.jpg", captured)
}

func TestCopyToClipboard_Unsupported(t *testing.T) {
	origUnsupported := ClipboardUnsupported
	defer func() { ClipboardUnsupported = origUnsupported }()

	ClipboardUnsupported = true

	err := CopyToClipboard("https://example.com/meme.jpg")
	assert.ErrorIs(t, err, ErrClipboardUnsupported)
}

func TestOpenInBrowser(t *testing.T) {
	original := BrowserOpen
	defer func() { BrowserOpen = original }()

	var captured string
	BrowserOpen = func(url string) error {
		captured = url
		return nil
	}

	err := OpenInBrowser("https://example.com/meme.jpg")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/meme.jpg", captured)
}
