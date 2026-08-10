package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismailshak/transit/internal/tui"
)

// Password renders a single-line masked input and blocks until the user
// submits or backs out. On cancel (Esc, Ctrl+C) returns ErrCancelled. On
// submitting with nothing typed returns ErrNoInput.
func Password(ctx context.Context, title string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 60
	ti.Prompt = "" // Control the prompt ourselves
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	prompt := promptModel{
		textInput: ti,
		title:     title,
	}

	program := tea.NewProgram(prompt, tea.WithContext(ctx))
	m, err := program.Run()
	if err != nil {
		return "", err
	}

	prompt = m.(promptModel)

	if prompt.cancelled {
		return "", ErrCancelled
	}

	input := prompt.textInput.Value()

	if input == "" {
		return "", ErrNoInput
	}

	return input, nil
}

// -- Internal model for the prompt component --

type promptModel struct {
	textInput textinput.Model
	title     string
	// cancelled distinguishes "user backed out" from "submitted empty"
	cancelled bool
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	//nolint:gocritic // bubbletea Update always type-switches on tea.Msg; more cases land as we handle more message types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	builder := strings.Builder{}

	builder.WriteString(tui.PROMPT_SYMBOL_STYLE(tui.PROMPT_SYMBOL))
	builder.WriteString(" ")
	builder.WriteString(tui.PROMPT_TITLE_STYLE(m.title))
	builder.WriteString(" ")
	builder.WriteString(m.textInput.View())

	return builder.String()
}
