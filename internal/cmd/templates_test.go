package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const templatesListJSON = `[
	{"id":"drake","name":"Drake Hotline Bling","lines":2,"styles":["default","animated"]},
	{"id":"buzz","name":"Buzz Lightyear","lines":2,"styles":["default"]},
	{"id":"fry","name":"Futurama Fry","lines":2,"styles":["default","animated"]}
]`

const templateDetailJSON = `{
	"id":"drake",
	"name":"Drake Hotline Bling",
	"lines":2,
	"overlays":0,
	"styles":["default","animated"],
	"blank":"https://api.memegen.link/images/drake.png",
	"example":{"text":["top","bottom"],"url":"https://api.memegen.link/images/drake/top/bottom.png"},
	"source":"https://knowyourmeme.com/memes/drakeposting",
	"keywords":["drake","bling","hotline"],
	"_self":"https://api.memegen.link/templates/drake"
}`

func TestTemplatesCmd_List(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &TemplatesCmd{}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{NoInput: true}))
	})

	assert.Contains(t, output, "drake")
	assert.Contains(t, output, "Drake Hotline Bling")
	assert.Contains(t, output, "buzz")
	assert.Contains(t, output, "3 templates")
}

func TestTemplatesCmd_List_JSON(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &TemplatesCmd{}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &parsed))
	assert.Len(t, parsed, 3)
	assert.Equal(t, "drake", parsed[0]["id"])
}

func TestTemplatesCmd_List_Filter(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"drake","name":"Drake Hotline Bling","lines":2,"styles":["default"]}]`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &TemplatesCmd{Filter: "drake"}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{NoInput: true}))
	})

	assert.Contains(t, gotQuery, "filter=drake")
	assert.Contains(t, output, "1 templates")
}

func TestTemplatesCmd_List_Animated(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &TemplatesCmd{Animated: true}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{NoInput: true}))
	})

	// drake + fry are animated, buzz is not
	assert.Contains(t, output, "drake")
	assert.Contains(t, output, "fry")
	assert.NotContains(t, output, "buzz")
	assert.Contains(t, output, "2 templates")
}

func TestTemplatesCmd_Detail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/templates/drake", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templateDetailJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &TemplatesCmd{ID: "drake"}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	assert.Contains(t, output, "ID:       drake")
	assert.Contains(t, output, "Name:     Drake Hotline Bling")
	assert.Contains(t, output, "Lines:    2")
	assert.Contains(t, output, "default, animated")
	assert.Contains(t, output, "Keywords: drake, bling, hotline")
}

func TestTemplatesCmd_Detail_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templateDetailJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &TemplatesCmd{ID: "drake"}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx, &RootFlags{})
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(buf[:n], &parsed))
	assert.Equal(t, "drake", parsed["id"])
	assert.Equal(t, "Drake Hotline Bling", parsed["name"])
}

func TestTemplatesCmd_Detail_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"template 'xyz' not found"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &TemplatesCmd{ID: "xyz"}

	err := cmd.Run(ctx, &RootFlags{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTemplatesCmd_NoClient(t *testing.T) {
	cmd := &TemplatesCmd{}
	err := cmd.Run(testCtxNoClient(t, false), &RootFlags{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api client not found")
}

// --- Cache integration tests ---

func TestTemplatesCmd_List_UsesCache(t *testing.T) {
	// Pre-populate cache, use a server that counts requests.
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	seedTemplateCache(t, cacheDir)

	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtxWithConfig(t, srv.URL)
	cmd := &TemplatesCmd{}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	assert.Equal(t, 0, requestCount, "should not hit API when cache is fresh")
	assert.Contains(t, output, "drake")
	assert.Contains(t, output, "2 templates")
}

func TestTemplatesCmd_List_RefreshBypassesCache(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	seedTemplateCache(t, cacheDir)

	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtxWithConfig(t, srv.URL)
	cmd := &TemplatesCmd{Refresh: true}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	assert.Equal(t, 1, requestCount, "should hit API when --refresh")
	assert.Contains(t, output, "3 templates")
}

func TestTemplatesCmd_List_FilterBypassesCache(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	seedTemplateCache(t, cacheDir)

	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"drake","name":"Drake Hotline Bling","lines":2,"styles":["default"]}]`))
	}))
	defer srv.Close()

	ctx := testCtxWithConfig(t, srv.URL)
	cmd := &TemplatesCmd{Filter: "drake"}

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	assert.Equal(t, 1, requestCount, "should hit API when --filter")
	assert.Contains(t, output, "1 templates")
}

func TestTemplatesCmd_List_PopulatesCache(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesListJSON))
	}))
	defer srv.Close()

	ctx := testCtxWithConfig(t, srv.URL)
	cmd := &TemplatesCmd{}

	captureStdout(t, func() {
		require.NoError(t, cmd.Run(ctx, &RootFlags{}))
	})

	// Verify cache file was created.
	cachePath := filepath.Join(cacheDir, "memelink", "templates.json")
	_, err := os.Stat(cachePath)
	assert.NoError(t, err, "cache file should exist after API fetch")
}

// seedTemplateCache writes a valid cache file with known templates.
func seedTemplateCache(t *testing.T, cacheDir string) {
	t.Helper()

	dir := filepath.Join(cacheDir, "memelink")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	tc := struct {
		Templates []map[string]any `json:"templates"`
		FetchedAt string           `json:"fetched_at"`
	}{
		Templates: []map[string]any{
			{"id": "drake", "name": "Drake Hotline Bling", "lines": float64(2), "styles": []string{"default", "animated"}},
			{"id": "fry", "name": "Futurama Fry", "lines": float64(2), "styles": []string{"default"}},
		},
		FetchedAt: time.Now().Format(time.RFC3339Nano),
	}

	data, err := json.MarshalIndent(tc, "", "  ")
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "templates.json"), data, 0o644))
}
