// Package tui contains functions that print pretty output to the terminal.
//
// Generally encompasses functions that are transit's user interface, where visual aesthetic matters.
// Regular messaging should be deferred to the `cli`.
package tui

import "github.com/charmbracelet/lipgloss"

const (
	SuccessIcon  = "✔"
	ErrorIcon    = "✖"
	SkipIcon     = "✖"
	PromptSymbol = "?"
)

// TODO: Confirm color visibility in light themed terminals
// (better yet, create a color palette for both light and dark themes).
var (
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Orange    = lipgloss.AdaptiveColor{Light: "#FF5F00", Dark: "#FFAF00"}
	DefaultFG = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	Green     = lipgloss.AdaptiveColor{Light: "#B4BE82", Dark: "#B4BE82"}
	Cyan      = lipgloss.AdaptiveColor{Light: "#00A3CC", Dark: "#00A3CC"}
	Red       = lipgloss.AdaptiveColor{Light: "#FF2A00", Dark: "#FF2A00"}
	Purple    = lipgloss.AdaptiveColor{Light: "#A093C7", Dark: "#A093C7"}

	SpinnerStyle = lipgloss.NewStyle().Foreground(Purple)

	OpSuccessStyle = lipgloss.NewStyle().Foreground(Green).Render
	OpFailedStyle  = lipgloss.NewStyle().Foreground(Red).Render
	OpSkippedStyle = lipgloss.NewStyle().Foreground(Subtle).Render

	PromptTitleStyle  = lipgloss.NewStyle().Bold(true).Render
	PromptSymbolStyle = lipgloss.NewStyle().Foreground(Cyan).Render
)

func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}
