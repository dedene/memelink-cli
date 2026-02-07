package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/api"
	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
)

func testCtx(t *testing.T, baseURL string, jsonMode bool) context.Context {
	t.Helper()

	client := api.NewClient(api.ClientOptions{
		BaseURL:   baseURL,
		UserAgent: "memelink-cli/test",
	})

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: jsonMode})
	ctx = api.WithClient(ctx, client)

	return ctx
}

func testCtxNoClient(t *testing.T, jsonMode bool) context.Context {
	t.Helper()

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: jsonMode})

	return ctx
}

// testCtxWithConfig returns a context with API client and default config.
func testCtxWithConfig(t *testing.T, baseURL string, jsonMode bool) context.Context {
	t.Helper()

	ctx := testCtx(t, baseURL, jsonMode)
	ctx = config.WithConfig(ctx, &config.Config{})

	return ctx
}

// testCtxWithCfg returns a context with API client and the given config.
func testCtxWithCfg(t *testing.T, baseURL string, jsonMode bool, cfg *config.Config) context.Context {
	t.Helper()

	ctx := testCtx(t, baseURL, jsonMode)
	ctx = config.WithConfig(ctx, cfg)

	return ctx
}

// captureStdout runs fn while capturing os.Stdout and returns the output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = origStdout

	buf, _ := io.ReadAll(r)
	_ = r.Close()

	return string(buf)
}

// --- Auto-generate mode tests (from Plan 01) ---

func TestGenerateCmd_AutoGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/buzz/when_the_code_works.png","generator":"Pattern","confidence":0.46}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{Template: "when the code works"}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })

	require.NoError(t, runErr)
	assert.Contains(t, output, "https://api.memegen.link/images/buzz/when_the_code_works.png")
}

func TestGenerateCmd_AutoGenerate_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/buzz/test.png","generator":"Pattern","confidence":0.85}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &GenerateCmd{Template: "when it works"}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })

	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &parsed))
	assert.Equal(t, "https://api.memegen.link/images/buzz/test.png", parsed["url"])
	assert.Equal(t, "Pattern", parsed["generator"])
	assert.InDelta(t, 0.85, parsed["confidence"], 0.001)
}

func TestGenerateCmd_NoArgs(t *testing.T) {
	cmd := &GenerateCmd{}
	err := cmd.Run(context.Background(), &RootFlags{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide text")
}

func TestGenerateCmd_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"text is required"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{Template: "test"}

	err := cmd.Run(ctx, &RootFlags{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "text is required")
}

// --- Template mode tests ---

func TestGenerateCmd_TemplateMode(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/top/bottom.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"top", "bottom"},
		Format:   "jpg",
		Layout:   "default",
	}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })

	require.NoError(t, runErr)
	assert.Equal(t, "/images", gotPath)
	assert.Contains(t, output, "https://api.memegen.link/images/drake/top/bottom.jpg")
}

func TestGenerateCmd_TemplateMode_WithFlags(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.png"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
		Format:   "png",
		Font:     "impact",
		Layout:   "top",
		Style:    []string{"default"},
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "drake", parsed["template_id"])
	assert.Equal(t, []any{"a", "b"}, parsed["text"])
	assert.Equal(t, "png", parsed["extension"])
	assert.Equal(t, "impact", parsed["font"])
	assert.Equal(t, "top", parsed["layout"])
	assert.Equal(t, []any{"default"}, parsed["style"])
	assert.Equal(t, false, parsed["redirect"])
}

func TestGenerateCmd_TemplateMode_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template:  "drake",
		Text:      []string{"a", "b"},
		Format:    "jpg",
		Layout:    "default",
		Width:     400,
		Height:    300,
		TextColor: []string{"red", "blue"},
	}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	u, err := url.Parse(output[:len(output)-1]) // trim newline
	require.NoError(t, err)
	assert.Equal(t, "400", u.Query().Get("width"))
	assert.Equal(t, "300", u.Query().Get("height"))
	assert.Equal(t, "red,blue", u.Query().Get("color"))
}

func TestGenerateCmd_TemplateMode_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
		Format:   "jpg",
		Layout:   "default",
	}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &parsed))
	assert.Equal(t, "https://api.memegen.link/images/drake/a/b.jpg", parsed["url"])
	// Template mode should NOT have generator/confidence
	assert.Nil(t, parsed["generator"])
	assert.Nil(t, parsed["confidence"])
}

// --- Custom mode tests ---

func TestGenerateCmd_CustomMode(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/custom/hello.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template:   "custom",
		Text:       []string{"hello"},
		Background: "https://example.com/img.jpg",
		Format:     "jpg",
		Layout:     "default",
	}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })

	require.NoError(t, runErr)
	assert.Equal(t, "/images/custom", gotPath)
	assert.Contains(t, output, "https://api.memegen.link/images/custom/hello.jpg")
}

func TestGenerateCmd_CustomMode_NoBackground(t *testing.T) {
	cmd := &GenerateCmd{
		Template: "custom",
		Text:     []string{"hello"},
		Format:   "jpg",
		Layout:   "default",
	}
	ctx := testCtx(t, "http://unused", false)
	err := cmd.Run(ctx, &RootFlags{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--background required")
}

func TestGenerateCmd_CustomMode_StyleJoined(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/custom/hello.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template:   "custom",
		Text:       []string{"hello"},
		Background: "https://example.com/img.jpg",
		Style:      []string{"default", "animated"},
		Format:     "jpg",
		Layout:     "default",
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	// Style is joined as comma-separated string for custom endpoint
	assert.Equal(t, "default,animated", parsed["style"])
}

// --- Safe flag mode tests ---

func TestGenerateCmd_SafeFlag_AutoMode(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/buzz/test.png"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template: "test text",
		Safe:     true,
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	// Safe should be in POST body for auto mode
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, true, parsed["safe"])
}

func TestGenerateCmd_SafeFlag_TemplateMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
		Safe:     true,
		Format:   "jpg",
		Layout:   "default",
	}

	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	// Safe should be appended as query param for template mode
	u, err := url.Parse(output[:len(output)-1])
	require.NoError(t, err)
	assert.Equal(t, "true", u.Query().Get("safe"))
}

// --- Config defaults tests ---

func TestGenerateCmd_UsesConfigFormat(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.png"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{DefaultFormat: "png"}
	ctx := testCtxWithCfg(t, srv.URL, false, cfg)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "png", parsed["extension"], "should use config default_format")
}

func TestGenerateCmd_FlagOverridesConfig(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.gif"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{DefaultFormat: "png"}
	ctx := testCtxWithCfg(t, srv.URL, false, cfg)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
		Format:   "gif",
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "gif", parsed["extension"], "explicit flag overrides config")
}

func TestGenerateCmd_UsesConfigSafe(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/buzz/test.png"}`))
	}))
	defer srv.Close()

	safeTrue := true
	cfg := &config.Config{Safe: &safeTrue}
	ctx := testCtxWithCfg(t, srv.URL, false, cfg)
	cmd := &GenerateCmd{
		Template: "some safe text",
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, true, parsed["safe"], "config safe=true should apply without --safe flag")
}

func TestGenerateCmd_UsesConfigFont(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{DefaultFont: "impact"}
	ctx := testCtxWithCfg(t, srv.URL, false, cfg)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "impact", parsed["font"], "should use config default_font")
}

func TestGenerateCmd_UsesConfigLayout(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{DefaultLayout: "top"}
	ctx := testCtxWithCfg(t, srv.URL, false, cfg)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "top", parsed["layout"], "should use config default_layout")
}

func TestGenerateCmd_DefaultsWithoutConfig(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/a/b.jpg"}`))
	}))
	defer srv.Close()

	// No config in context at all.
	ctx := testCtx(t, srv.URL, false)
	cmd := &GenerateCmd{
		Template: "drake",
		Text:     []string{"a", "b"},
	}

	var runErr error
	captureStdout(t, func() { runErr = cmd.Run(ctx, &RootFlags{}) })
	require.NoError(t, runErr)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "jpg", parsed["extension"], "hardcoded default format is jpg")
	assert.Equal(t, "default", parsed["layout"], "hardcoded default layout is default")
}
