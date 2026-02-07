package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fontsListJSON = `[
	{"id":"impact","alias":"impact-alias","filename":"impact.ttf","_self":"https://api.memegen.link/fonts/impact"},
	{"id":"arial","alias":null,"filename":"arial.ttf","_self":"https://api.memegen.link/fonts/arial"}
]`

const fontDetailJSON = `{
	"id":"impact",
	"alias":"impact-alias",
	"filename":"impact.ttf",
	"_self":"https://api.memegen.link/fonts/impact"
}`

const fontDetailNilAliasJSON = `{
	"id":"arial",
	"alias":null,
	"filename":"arial.ttf",
	"_self":"https://api.memegen.link/fonts/arial"
}`

func TestFontsCmd_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontsListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &FontsCmd{}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()

	output := string(buf[:n])
	assert.Contains(t, output, "impact")
	assert.Contains(t, output, "arial")
	assert.Contains(t, output, "2 fonts")
}

func TestFontsCmd_List_NilAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontsListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &FontsCmd{}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()

	// arial has null alias, should show "-"
	output := string(buf[:n])
	assert.Contains(t, output, "-")
}

func TestFontsCmd_List_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontsListJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &FontsCmd{}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal(buf[:n], &parsed))
	assert.Len(t, parsed, 2)
	assert.Equal(t, "impact", parsed[0]["id"])
}

func TestFontsCmd_Detail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/fonts/impact", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontDetailJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &FontsCmd{ID: "impact"}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()

	output := string(buf[:n])
	assert.Contains(t, output, "ID:       impact")
	assert.Contains(t, output, "Alias:    impact-alias")
	assert.Contains(t, output, "Filename: impact.ttf")
}

func TestFontsCmd_Detail_NilAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontDetailNilAliasJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &FontsCmd{ID: "arial"}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()

	output := string(buf[:n])
	assert.Contains(t, output, "Alias:    -")
}

func TestFontsCmd_Detail_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fontDetailJSON))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, true)
	cmd := &FontsCmd{ID: "impact"}

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w

	runErr := cmd.Run(ctx)
	_ = w.Close()
	os.Stdout = origStdout

	require.NoError(t, runErr)

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(buf[:n], &parsed))
	assert.Equal(t, "impact", parsed["id"])
	assert.Equal(t, "impact-alias", parsed["alias"])
}

func TestFontsCmd_Detail_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"font 'xyz' not found"}`))
	}))
	defer srv.Close()

	ctx := testCtx(t, srv.URL, false)
	cmd := &FontsCmd{ID: "xyz"}

	err := cmd.Run(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFontsCmd_NoClient(t *testing.T) {
	cmd := &FontsCmd{}
	err := cmd.Run(testCtxNoClient(t, false))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api client not found")
}
