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
// CallbackFn returns and the spinner has cleaned up after itself. Cancelling
// (Esc, Ctrl+C) returns ErrCancelled, same as Password — it only stops the
// animation, opts.CallbackFn keeps running to completion regardless.
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
	// CallbackFn runs on this one. Buffered so the goroutine's send
	// never blocks on us reading it below.
	modelCh := make(chan tea.Model, 1)
	go func() {
		m, err := program.Run()
		if err != nil {
			// TODO: Handle this better? Maybe send via channel and errors.Join
			logger.Debug(err)
		}
		modelCh <- m
	}()

	err := opts.CallbackFn()

	// Quit signals the program to stop.
	// Wait blocks until it actually stops so the
	// terminal's restored before we print anything below
	program.Quit()
	program.Wait()

	sp = (<-modelCh).(spinnerModel)
	if sp.cancelled {
		return ErrCancelled
	}

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
	// cancelled distinguishes "user backed out" from "CallbackFn finished"
	cancelled bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	//nolint:gocritic // bubbletea Update always type-switches on tea.Msg; more cases land as we handle more message types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		//nolint:gocritic
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
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
