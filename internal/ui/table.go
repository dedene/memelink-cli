package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// RenderTable builds a formatted table string using lipgloss.
// When color is true, headers are styled and borders are rendered with color.
// When color is false, a plain ASCII table is produced.
func RenderTable(headers []string, rows [][]string, color bool) string {
	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderRow(false).
		BorderColumn(true).
		BorderHeader(true)

	if color {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7c3aed"))
		cellStyle := lipgloss.NewStyle()

		t.StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		})
	}

	return t.Render()
}
