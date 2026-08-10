package ui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/tui"
)

type SpinnerOptions struct {
	// SpinMessage is shown next to the spinner while CallbackFn is running
	SpinMessage string
	// SuccessMessage replaces the spinner once CallbackFn returns nil
	SuccessMessage string
	// ErrorMessage replaces the spinner once CallbackFn returns an error
	ErrorMessage string
	// CallbackFn is the work to run while the spinner animates
	CallbackFn func() error
}

// WithSpinner animates a spinner while opts.CallbackFn runs, then swaps it
// for a success or error message depending on the result. Blocks until
// CallbackFn returns and the spinner has cleaned up after itself.
func WithSpinner(ctx context.Context, opts *SpinnerOptions) error {
	sp := spinnerModel{
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(tui.SPINNER_STYLE),
		),
		msg: &opts.SpinMessage,
	}

	program := tea.NewProgram(
		sp,
		tea.WithContext(ctx),
	)

	// Run the spinner on its own goroutine so it can animate while
	// CallbackFn runs on this one
	go func() {
		_, err := program.Run()
		if err != nil {
			// TODO: Handle this better? Maybe send via channel and errors.Join
			logger.Debug(err)
		}
	}()

	err := opts.CallbackFn()

	// Quit signals the program to stop.
	// Wait blocks until it actually stops so the
	// terminal's restored before we print anything below
	program.Quit()
	program.Wait()

	if err != nil {
		tui.OperationFailed(opts.ErrorMessage)

		return err
	}

	tui.OperationSuccessful(opts.SuccessMessage)

	return nil
}

// -- Internal model for the spinner component --

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
