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
	GREEN      = lipgloss.AdaptiveColor{Light: "#A093C7", Dark: "#A093C7"}
	CYAN       = lipgloss.AdaptiveColor{Light: "#00A3CC", Dark: "#00A3CC"}

	// TODO: Confirm color visibility in light themed terminals
	// (better yet, create a color palette for both light and dark themes)
	SPINNER_STYLE = lipgloss.NewStyle().Foreground(lipgloss.Color("#A093C7"))

	OP_SUCCESS_STYLE = lipgloss.NewStyle().Foreground(lipgloss.Color("#B4BE82")).Render
	OP_FAILED_STYLE  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A00")).Render
	OP_SKIPPED_STYLE = lipgloss.NewStyle().Foreground(SUBTLE).Render

	PROMPT_TITLE_STYLE  = lipgloss.NewStyle().Bold(true).Render
	PROMPT_SYMBOL_STYLE = lipgloss.NewStyle().Foreground(CYAN).Render
)

const (
	SUCCESS_ICON  = "✔"
	ERROR_ICON    = "✖"
	SKIP_ICON     = "✖"
	PROMPT_SYMBOL = "?"
)

func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}
