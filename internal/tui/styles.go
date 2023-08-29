// Package tui contains functions that print pretty output to the terminal.
//
// Generally encompasses functions that are transit's user interface, where visual aesthetic matters.
// Regular messaging should be deferred to the `logger` package
package tui

import "github.com/charmbracelet/lipgloss"

var (
	SUBTLE     = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	ORANGE     = lipgloss.AdaptiveColor{Light: "#FF5F00", Dark: "#FFAF00"}
	DEFAULT_FG = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
)

func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}
