package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Prompt struct {
	prompt promptModel
}

func NewPasswordPrompt(title string) *Prompt {
	p := newPromptModel("", title)

	p.textInput.EchoMode = textinput.EchoPassword
	p.textInput.EchoCharacter = '•'

	return &Prompt{
		prompt: p,
	}
}

func NewPrompt(placeholder, title string) *Prompt {
	p := newPromptModel(placeholder, title)
	return &Prompt{
		prompt: p,
	}
}

func (p *Prompt) Render() string {
	m, _ := tea.NewProgram(p.prompt).Run() // TODO: handle error
	if m.(promptModel).cancelled {
		return ""
	}

	return m.(promptModel).textInput.Value()
}

// Internal model for the prompt component

type promptModel struct {
	textInput textinput.Model
	title     string
	cancelled bool
}

func newPromptModel(placeholder, title string) promptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 60
	ti.Prompt = "" // Control the prompt ourselves

	return promptModel{
		textInput: ti,
		title:     title,
	}
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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

	builder.WriteString(PROMPT_SYMBOL_STYLE(PROMPT_SYMBOL))
	builder.WriteString(" ")
	builder.WriteString(PROMPT_TITLE_STYLE(m.title))
	builder.WriteString(" ")
	builder.WriteString(m.textInput.View())

	return builder.String()
}
