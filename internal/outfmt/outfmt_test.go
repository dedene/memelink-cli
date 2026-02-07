package outfmt_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/outfmt"
)

func TestWithMode_IsJSON_RoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: true})
	assert.True(t, outfmt.IsJSON(ctx))
}

func TestIsJSON_BareContext(t *testing.T) {
	ctx := context.Background()
	assert.False(t, outfmt.IsJSON(ctx))
}

func TestWithMode_NotJSON(t *testing.T) {
	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, outfmt.Mode{JSON: false})
	assert.False(t, outfmt.IsJSON(ctx))
}

func TestWriteJSON_PrettyPrinted(t *testing.T) {
	var buf bytes.Buffer
	err := outfmt.WriteJSON(&buf, map[string]string{"hello": "world"})
	require.NoError(t, err)

	want := "{\n  \"hello\": \"world\"\n}\n"
	assert.Equal(t, want, buf.String())
}

func TestWriteJSON_NoHTMLEscaping(t *testing.T) {
	var buf bytes.Buffer
	err := outfmt.WriteJSON(&buf, map[string]string{
		"url": "https://example.com?a=1&b=2",
		"tag": "<b>bold</b>",
	})
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "&")
	assert.NotContains(t, out, "\\u0026")
	assert.Contains(t, out, "<b>bold</b>")
	assert.NotContains(t, out, "\\u003c")
	assert.NotContains(t, out, "\\u003e")
}
