package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTable_NoColor(t *testing.T) {
	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"drake", "Drake Hotline Bling"},
		{"buzz", "Buzz Lightyear"},
	}

	out := RenderTable(headers, rows, false)
	assert.Contains(t, out, "drake")
	assert.Contains(t, out, "Drake Hotline Bling")
	assert.Contains(t, out, "buzz")
	assert.Contains(t, out, "Buzz Lightyear")
	// Headers present
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "Name")
}

func TestRenderTable_WithColor(t *testing.T) {
	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"drake", "Drake Hotline Bling"},
	}

	out := RenderTable(headers, rows, true)
	// Should still contain data
	assert.Contains(t, out, "drake")
	assert.Contains(t, out, "Drake Hotline Bling")
}

func TestRenderTable_Empty(t *testing.T) {
	headers := []string{"ID", "Name"}
	out := RenderTable(headers, nil, false)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "Name")
}

func TestRenderTable_HasBorders(t *testing.T) {
	headers := []string{"Col"}
	rows := [][]string{{"val"}}

	out := RenderTable(headers, rows, false)
	// Should have border characters
	assert.True(t, strings.Contains(out, "│") || strings.Contains(out, "|"),
		"expected border character in output")
}
