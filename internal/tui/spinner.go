package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismailshak/transit/internal/logger"
)

type Spinner struct {
	spinner *spinnerModel
	program *tea.Program
}

// Create a new spinner with a message to display while spinning
func NewSpinner(msg string) *Spinner {
	m := spinner.New()
	m.Spinner = spinner.Dot
	m.Style = SPINNER_STYLE

	spinner := &spinnerModel{
		spinner: m,
		msg:     &msg,
	}

	program := tea.NewProgram(spinner)

	return &Spinner{spinner: spinner, program: program}
}

// Begin the spinner animation.
// NOTE: This function is blocking and you should call it in a goroutine
// if you want to execute other code while the spinner is running. Use
// the `Stop` function to stop the spinner.
func (s *Spinner) Start() {
	s.program.Run()
}

// Stop and clear the spinner from the terminal
func (s *Spinner) Stop() {
	s.program.Quit()
	s.program.Wait()
}

// Stop and replace the spinner with an error message
func (s *Spinner) Error(msg string) {
	s.program.Quit()
	s.program.Wait()
	icon := OP_FAILED_STYLE(ERROR_ICON)
	logger.Print(icon, msg)
}

// Stop and replace the spinner with a success message
func (s *Spinner) Success(msg string) {
	s.program.Quit()
	s.program.Wait()
	icon := OP_SUCCESS_STYLE(SUCCESS_ICON)
	logger.Print(icon, msg)
}

//
// Internal model for the spinner component
//

type spinnerModel struct {
	spinner spinner.Model
	msg     *string
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	// TODO: maybe we don't concatenate the message here? (runs every frame)
	return m.spinner.View() + *m.msg
}
