package ui_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/ui"
)

func TestNew_ValidColorModes(t *testing.T) {
	for _, mode := range []string{"auto", "always", "never", "", "  Auto ", "NEVER"} {
		t.Run(mode, func(t *testing.T) {
			u, err := ui.New(ui.Options{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Color:  mode,
			})
			require.NoError(t, err)
			assert.NotNil(t, u)
			assert.NotNil(t, u.Out())
			assert.NotNil(t, u.Err())
		})
	}
}

func TestNew_InvalidColorMode(t *testing.T) {
	_, err := ui.New(ui.Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Color:  "rainbow",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ui.ErrInvalidColor))
	assert.Contains(t, err.Error(), "rainbow")
}

func TestPrinter_Println(t *testing.T) {
	var buf bytes.Buffer
	u, err := ui.New(ui.Options{
		Stdout: &buf,
		Stderr: &bytes.Buffer{},
		Color:  "never",
	})
	require.NoError(t, err)

	u.Out().Println("hello")
	assert.Equal(t, "hello\n", buf.String())
}

func TestPrinter_Printf(t *testing.T) {
	var buf bytes.Buffer
	u, err := ui.New(ui.Options{
		Stdout: &buf,
		Stderr: &bytes.Buffer{},
		Color:  "never",
	})
	require.NoError(t, err)

	u.Out().Printf("count: %d", 42)
	assert.Equal(t, "count: 42\n", buf.String())
}

func TestPrinter_Errorf(t *testing.T) {
	var buf bytes.Buffer
	u, err := ui.New(ui.Options{
		Stdout: &bytes.Buffer{},
		Stderr: &buf,
		Color:  "never",
	})
	require.NoError(t, err)

	u.Err().Errorf("not found: %s", "drake")
	assert.Equal(t, "Error: not found: drake\n", buf.String())
}

func TestPrinter_NeverMode_NoColor(t *testing.T) {
	var buf bytes.Buffer
	u, err := ui.New(ui.Options{
		Stdout: &buf,
		Stderr: &bytes.Buffer{},
		Color:  "never",
	})
	require.NoError(t, err)

	assert.False(t, u.Out().ColorEnabled())
}

func TestWithUI_FromContext_RoundTrip(t *testing.T) {
	u, err := ui.New(ui.Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Color:  "never",
	})
	require.NoError(t, err)

	ctx := context.Background()
	ctx = ui.WithUI(ctx, u)
	got := ui.FromContext(ctx)
	assert.Same(t, u, got)
}

func TestFromContext_BareContext_Nil(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ui.FromContext(ctx))
}
