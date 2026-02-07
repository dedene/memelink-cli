package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/api"
)

var testTemplates = []api.Template{
	{ID: "drake", Name: "Drake Hotline Bling", Lines: 2, Styles: []string{"default", "animated"}},
	{ID: "fry", Name: "Futurama Fry", Lines: 2, Styles: []string{"default"}},
}

func TestLoadMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")

	templates, err := LoadTemplates(path, 24*time.Hour)
	require.NoError(t, err)
	assert.Nil(t, templates)
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")

	require.NoError(t, SaveTemplates(path, testTemplates))

	loaded, err := LoadTemplates(path, 24*time.Hour)
	require.NoError(t, err)
	require.Len(t, loaded, 2)
	assert.Equal(t, "drake", loaded[0].ID)
	assert.Equal(t, "Drake Hotline Bling", loaded[0].Name)
	assert.Equal(t, 2, loaded[0].Lines)
	assert.Equal(t, []string{"default", "animated"}, loaded[0].Styles)
	assert.Equal(t, "fry", loaded[1].ID)
}

func TestLoadExpired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")

	// Write cache with FetchedAt in the past.
	tc := TemplateCache{
		Templates: testTemplates,
		FetchedAt: time.Now().Add(-48 * time.Hour),
	}

	data, err := json.MarshalIndent(tc, "", "  ")
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(path, data, 0o644))

	loaded, loadErr := LoadTemplates(path, 24*time.Hour)
	require.NoError(t, loadErr)
	assert.Nil(t, loaded)
}

func TestLoadCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")

	require.NoError(t, os.WriteFile(path, []byte("{{{not json"), 0o644))

	loaded, err := LoadTemplates(path, 24*time.Hour)
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestLoadFresh(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")

	require.NoError(t, SaveTemplates(path, testTemplates))

	loaded, err := LoadTemplates(path, 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Len(t, loaded, 2)
}

func TestSaveCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deep")
	path := filepath.Join(dir, "templates.json")

	require.NoError(t, SaveTemplates(path, testTemplates))

	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestLoadZeroTTL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "templates.json")

	require.NoError(t, SaveTemplates(path, testTemplates))

	// 0 TTL means always expired.
	loaded, err := LoadTemplates(path, 0)
	require.NoError(t, err)
	assert.Nil(t, loaded)
}
