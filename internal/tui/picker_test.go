package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dedene/memelink-cli/internal/api"
)

func testItems() []list.Item {
	return []list.Item{
		NewTemplateItem(api.Template{
			ID:       "drake",
			Name:     "Drake Hotline Bling",
			Lines:    2,
			Keywords: []string{"drake", "no", "yes"},
			Example: struct {
				Text []string `json:"text"`
				URL  string   `json:"url"`
			}{Text: []string{"no", "yes"}},
		}),
		NewTemplateItem(api.Template{
			ID:       "fry",
			Name:     "Futurama Fry",
			Lines:    1,
			Keywords: []string{"fry", "not sure"},
		}),
	}
}

func testItemsWithZeroLines() []list.Item {
	return []list.Item{
		NewTemplateItem(api.Template{
			ID:    "noline",
			Name:  "No Lines Template",
			Lines: 0,
		}),
	}
}

func sizeMsg() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: 80, Height: 24}
}

func readyModel(t *testing.T) Model {
	t.Helper()

	m := NewPicker(testItems())
	result, _ := m.Update(sizeMsg())

	model, ok := result.(Model)
	require.True(t, ok)

	return model
}

func TestNewPicker_InitialState(t *testing.T) {
	m := NewPicker(testItems())

	assert.Equal(t, StatePicking, m.State())
	assert.False(t, m.Cancelled())
	assert.Nil(t, m.Selected())
	assert.False(t, m.ready)
}

func TestPicker_WindowSizeMsg(t *testing.T) {
	m := NewPicker(testItems())

	result, _ := m.Update(sizeMsg())
	model := result.(Model)

	assert.True(t, model.ready)
	assert.Equal(t, 80, model.width)
	assert.Equal(t, 24, model.height)
}

func TestPicker_EnterTransitionsToInputting(t *testing.T) {
	m := readyModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	assert.Equal(t, StateInputting, model.State())
	require.NotNil(t, model.Selected())
	assert.Equal(t, "drake", model.Selected().ID)
	assert.Len(t, model.inputs, 2)
}

func TestPicker_CtrlCCancels(t *testing.T) {
	m := readyModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := result.(Model)

	assert.Equal(t, StateDone, model.State())
	assert.True(t, model.Cancelled())
	assert.Nil(t, model.Selected())
}

func TestPicker_EscCancels(t *testing.T) {
	m := readyModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	assert.Equal(t, StateDone, model.State())
	assert.True(t, model.Cancelled())
	assert.Nil(t, model.Selected())
}

func TestPicker_ViewLoading(t *testing.T) {
	m := NewPicker(testItems())

	assert.Equal(t, "Loading...", m.View())
}

func TestPicker_ViewAfterReady(t *testing.T) {
	m := readyModel(t)

	view := m.View()
	assert.NotEmpty(t, view)
	assert.NotEqual(t, "Loading...", view)
}

func TestTemplateItem_Title(t *testing.T) {
	item := NewTemplateItem(api.Template{Name: "Drake Hotline Bling"})
	assert.Equal(t, "Drake Hotline Bling", item.Title())
}

func TestTemplateItem_Description(t *testing.T) {
	item := NewTemplateItem(api.Template{ID: "drake", Lines: 2})
	assert.Equal(t, "ID: drake | 2 lines", item.Description())
}

func TestTemplateItem_FilterValue(t *testing.T) {
	item := NewTemplateItem(api.Template{
		Name:     "Drake Hotline Bling",
		ID:       "drake",
		Keywords: []string{"drake", "no", "yes"},
	})
	assert.Equal(t, "Drake Hotline Bling drake drake no yes", item.FilterValue())
}

func TestTemplateItem_Template(t *testing.T) {
	tmpl := api.Template{ID: "drake", Name: "Drake Hotline Bling", Lines: 2}
	item := NewTemplateItem(tmpl)
	assert.Equal(t, tmpl, item.Template())
}

// --- Text Input (StateInputting) Tests ---

// inputtingModel returns a model transitioned to StateInputting via Enter on drake (2 lines).
func inputtingModel(t *testing.T) Model {
	t.Helper()

	m := readyModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, ok := result.(Model)
	require.True(t, ok)
	require.Equal(t, StateInputting, model.State())

	return model
}

func TestPicker_ZeroLines_SkipsInput(t *testing.T) {
	items := testItemsWithZeroLines()
	m := NewPicker(items)

	// Send size to make ready.
	result, _ := m.Update(sizeMsg())
	model := result.(Model)

	// Enter on 0-line template.
	result, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = result.(Model)

	assert.Equal(t, StateDone, model.State())
	require.NotNil(t, model.Selected())
	assert.Equal(t, "noline", model.Selected().ID)
	assert.Equal(t, []string{}, model.Texts())
}

func TestInputting_EnterAdvancesFocus(t *testing.T) {
	m := inputtingModel(t)

	// First input should be focused.
	assert.Equal(t, 0, m.focusIdx)

	// Enter advances to second input.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	assert.Equal(t, StateInputting, model.State())
	assert.Equal(t, 1, model.focusIdx)
}

func TestInputting_EnterOnLastConfirms(t *testing.T) {
	m := inputtingModel(t)

	// Type in first input.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")})
	m = result.(Model)

	// Advance to second.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = result.(Model)

	// Type in second input.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("world")})
	m = result.(Model)

	// Enter on last input -> StateDone.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := result.(Model)

	assert.Equal(t, StateDone, model.State())
	assert.False(t, model.Cancelled())
	require.Len(t, model.Texts(), 2)
	assert.Equal(t, "hello", model.Texts()[0])
	assert.Equal(t, "world", model.Texts()[1])
}

func TestInputting_EscReturnsToPickig(t *testing.T) {
	m := inputtingModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := result.(Model)

	assert.Equal(t, StatePicking, model.State())
	assert.Nil(t, model.inputs)
}

func TestInputting_TabNavigatesForward(t *testing.T) {
	m := inputtingModel(t)

	assert.Equal(t, 0, m.focusIdx)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := result.(Model)

	assert.Equal(t, 1, model.focusIdx)
}

func TestInputting_ShiftTabNavigatesBack(t *testing.T) {
	m := inputtingModel(t)

	// Advance to second.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(Model)
	assert.Equal(t, 1, m.focusIdx)

	// Shift+tab back to first.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model := result.(Model)

	assert.Equal(t, 0, model.focusIdx)
}

func TestInputting_ShiftTabAtFirstDoesNothing(t *testing.T) {
	m := inputtingModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model := result.(Model)

	assert.Equal(t, 0, model.focusIdx)
}

func TestInputting_TabAtLastDoesNothing(t *testing.T) {
	m := inputtingModel(t)

	// Advance to last (idx 1 of 2).
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = result.(Model)
	assert.Equal(t, 1, m.focusIdx)

	// Tab again -- should stay at 1.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := result.(Model)

	assert.Equal(t, 1, model.focusIdx)
}

func TestInputting_CtrlCCancels(t *testing.T) {
	m := inputtingModel(t)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := result.(Model)

	assert.Equal(t, StateDone, model.State())
	assert.True(t, model.Cancelled())
}

func TestInputting_PlaceholdersFromExample(t *testing.T) {
	m := inputtingModel(t)

	assert.Equal(t, "no", m.inputs[0].Placeholder)
	assert.Equal(t, "yes", m.inputs[1].Placeholder)
}

func TestInputting_ViewContainsTemplateName(t *testing.T) {
	m := inputtingModel(t)

	view := m.View()
	assert.Contains(t, view, "Template: Drake Hotline Bling")
	assert.Contains(t, view, "Line 1:")
	assert.Contains(t, view, "Line 2:")
}

func TestTexts_EmptyBeforeConfirm(t *testing.T) {
	m := inputtingModel(t)
	assert.Nil(t, m.Texts())
}
