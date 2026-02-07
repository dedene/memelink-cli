package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dedene/memelink-cli/internal/api"
)

// State represents the current phase of the TUI model.
type State int

const (
	// StatePicking is the fuzzy template picker phase.
	StatePicking State = iota
	// StateInputting is the text input phase (used by plan 02).
	StateInputting
	// StateDone means the TUI is finished and ready to quit.
	StateDone
)

// Model is the bubbletea model for the template picker TUI.
type Model struct {
	state     State
	list      list.Model
	selected  *api.Template
	cancelled bool
	width     int
	height    int
	ready     bool

	// Text input fields (stateInputting).
	inputs   []textinput.Model
	focusIdx int
	texts    []string
}

// NewPicker creates a new picker Model with the given list items.
func NewPicker(items []list.Item) Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a template"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	return Model{
		state: StatePicking,
		list:  l,
	}
}

// Init returns the initial command. The list handles its own init internally.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize globally.
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		m.list.SetSize(wsm.Width, wsm.Height-2)

		for i := range m.inputs {
			m.inputs[i].Width = wsm.Width - 4
		}

		m.ready = true

		return m, nil
	}

	// Dispatch by state.
	switch m.state {
	case StatePicking:
		return m.updatePicking(msg)
	case StateInputting:
		return m.updateInputting(msg)
	}

	return m, nil
}

// updatePicking handles messages in the template picker state.
func (m Model) updatePicking(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			m.cancelled = true
			m.state = StateDone

			return m, tea.Quit

		case "esc":
			// Only quit on esc when not actively filtering.
			if m.list.FilterState() != list.Filtering {
				m.cancelled = true
				m.state = StateDone

				return m, tea.Quit
			}

		case "enter":
			// When actively filtering, delegate to list (confirms filter).
			if m.list.FilterState() == list.Filtering {
				break // fall through to list.Update
			}

			return m.handlePickEnter()
		}
	}

	// Delegate to list component.
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

// View renders the current TUI state.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	switch m.state {
	case StatePicking:
		return m.list.View()
	case StateInputting:
		return m.viewInputting()
	}

	return ""
}

// Selected returns the selected template, or nil if none selected.
func (m Model) Selected() *api.Template { return m.selected }

// Cancelled returns true if the user cancelled the picker.
func (m Model) Cancelled() bool { return m.cancelled }

// State returns the current picker state.
func (m Model) State() State { return m.state }

// Texts returns the collected text input values after confirmation.
func (m Model) Texts() []string { return m.texts }

// handlePickEnter processes Enter in statePicking: selects template and
// transitions to stateInputting (or StateDone for 0-line templates).
func (m Model) handlePickEnter() (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(TemplateItem)
	if !ok {
		return m, nil
	}

	t := item.Template()
	m.selected = &t

	// Templates with 0 lines skip text input.
	if t.Lines == 0 {
		m.texts = []string{}
		m.state = StateDone

		return m, tea.Quit
	}

	// Create text inputs for each line.
	m.inputs = make([]textinput.Model, t.Lines)

	for i := range t.Lines {
		ti := textinput.New()

		// Use example text as placeholder if available.
		if i < len(t.Example.Text) {
			ti.Placeholder = t.Example.Text[i]
		} else {
			ti.Placeholder = fmt.Sprintf("Line %d", i+1)
		}

		ti.CharLimit = 200

		if m.width > 4 {
			ti.Width = m.width - 4
		}

		m.inputs[i] = ti
	}

	m.inputs[0].Focus()
	m.focusIdx = 0
	m.state = StateInputting

	return m, textinput.Blink
}

// updateInputting handles messages in the text input state.
func (m Model) updateInputting(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		// Delegate non-key messages to focused input.
		var cmd tea.Cmd
		m.inputs[m.focusIdx], cmd = m.inputs[m.focusIdx].Update(msg)

		return m, cmd
	}

	switch keyMsg.String() {
	case "ctrl+c":
		m.cancelled = true
		m.state = StateDone

		return m, tea.Quit

	case "esc":
		// Go back to picker.
		m.state = StatePicking
		m.inputs = nil
		m.focusIdx = 0

		return m, nil

	case "enter":
		if m.focusIdx < len(m.inputs)-1 {
			// Advance to next input.
			m.inputs[m.focusIdx].Blur()
			m.focusIdx++
			m.inputs[m.focusIdx].Focus()

			return m, textinput.Blink
		}

		// Last input -- collect and finish.
		m.texts = make([]string, len(m.inputs))
		for i := range m.inputs {
			m.texts[i] = m.inputs[i].Value()
		}

		m.state = StateDone

		return m, tea.Quit

	case "tab":
		if m.focusIdx < len(m.inputs)-1 {
			m.inputs[m.focusIdx].Blur()
			m.focusIdx++
			m.inputs[m.focusIdx].Focus()

			return m, textinput.Blink
		}

		return m, nil

	case "shift+tab":
		if m.focusIdx > 0 {
			m.inputs[m.focusIdx].Blur()
			m.focusIdx--
			m.inputs[m.focusIdx].Focus()

			return m, textinput.Blink
		}

		return m, nil
	}

	// Delegate to focused input for typing.
	var cmd tea.Cmd
	m.inputs[m.focusIdx], cmd = m.inputs[m.focusIdx].Update(msg)

	return m, cmd
}

// viewInputting renders the text input form.
func (m Model) viewInputting() string {
	name := ""
	if m.selected != nil {
		name = m.selected.Name
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Template: %s\n\n", name)

	for i, input := range m.inputs {
		fmt.Fprintf(&b, "  Line %d: %s\n", i+1, input.View())
	}

	b.WriteString("\n  Enter: next/confirm | Esc: back | Ctrl+C: quit\n")

	return b.String()
}
