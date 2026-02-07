package preview

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// tiny1x1PNG generates a valid 1x1 red PNG in memory.
func tiny1x1PNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}

func TestShow_Success(t *testing.T) {
	data := tiny1x1PNG(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	}))
	defer srv.Close()

	var out bytes.Buffer
	err := Show(context.Background(), srv.URL, Options{
		Width:  40,
		Writer: &out,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, out.Bytes(), "expected rendered output")
}

func TestShow_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var out bytes.Buffer
	err := Show(context.Background(), srv.URL, Options{
		Width:  40,
		Writer: &out,
	})

	assert.NoError(t, err)
	assert.Empty(t, out.Bytes(), "expected no output on HTTP error")
}

func TestShow_InvalidURL(t *testing.T) {
	var out bytes.Buffer
	err := Show(context.Background(), "://bad-url", Options{
		Width:  40,
		Writer: &out,
	})

	assert.NoError(t, err)
	assert.Empty(t, out.Bytes(), "expected no output on invalid URL")
}

func TestShow_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(tiny1x1PNG(t))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var out bytes.Buffer
	err := Show(ctx, srv.URL, Options{
		Width:  40,
		Writer: &out,
	})

	assert.NoError(t, err)
	assert.Empty(t, out.Bytes(), "expected no output on cancelled context")
}
