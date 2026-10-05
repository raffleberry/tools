package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func EditText(initialText string) (string, bool) {
	ti := textarea.New()
	ti.SetValue(initialText)
	ti.Placeholder = "Write your message here..."
	ti.ShowLineNumbers = false
	ti.Focus()

	m := model{
		textarea: ti,
		isSaving: false,
		saved:    false,
	}

	// Enable the alternate screen buffer so it feels like a true full-screen app (like Vim/Nano)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return initialText, false
	}

	if m, ok := finalModel.(model); ok {
		return m.textarea.Value(), m.saved
	}

	return initialText, false
}

type model struct {
	textarea textarea.Model
	isSaving bool
	saved    bool
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// 1. Capture terminal size updates dynamically
	case tea.WindowSizeMsg:
		// Reserve height lines for your headers/footers so the text doesn't clip
		// 3 lines for header, 2 lines for spacing/borders
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height - 5)

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		if m.isSaving {
			switch msg.Type {
			case tea.KeyEsc:
				m.isSaving = false
				m.textarea.Focus()
				return m, nil
			case tea.KeyEnter:
				m.saved = true
				return m, tea.Quit
			}
		} else {
			switch msg.Type {
			case tea.KeyCtrlS:
				m.isSaving = true
				m.textarea.Blur()
				return m, nil
			}
		}
	}

	if !m.isSaving {
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.isSaving {
		return "\n  Are you sure you want to save?\n\n  [Enter] Save and Exit\n  [Esc]   Keep Editing\n"
	}

	header := lipgloss.NewStyle().
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1).
		Render(" EDITOR ")

	subHeader := "  Ctrl+S: Save | Ctrl+C: Cancel\n"

	return fmt.Sprintf("%s\n%s\n%s", header, subHeader, m.textarea.View())
}
