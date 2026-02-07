// Package tui provides the interactive Bubbletea TUI for template selection and meme generation.
package tui

import (
	"fmt"
	"strings"

	"github.com/dedene/memelink-cli/internal/api"
)

// TemplateItem wraps api.Template to implement the bubbles list.DefaultItem
// interface. It provides Title, Description, and FilterValue for the fuzzy
// picker list component.
type TemplateItem struct {
	template api.Template
}

// NewTemplateItem creates a TemplateItem from an api.Template.
func NewTemplateItem(t api.Template) TemplateItem {
	return TemplateItem{template: t}
}

// Title returns the template name for list display.
func (i TemplateItem) Title() string { return i.template.Name }

// Description returns template ID and line count for list display.
func (i TemplateItem) Description() string {
	return fmt.Sprintf("ID: %s | %d lines", i.template.ID, i.template.Lines)
}

// FilterValue returns a combined string of name, ID, and keywords for fuzzy matching.
func (i TemplateItem) FilterValue() string {
	return i.template.Name + " " + i.template.ID + " " + strings.Join(i.template.Keywords, " ")
}

// Template returns the wrapped api.Template.
func (i TemplateItem) Template() api.Template { return i.template }
