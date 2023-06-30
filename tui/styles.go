package tui

import "github.com/charmbracelet/lipgloss"

var (
	SUBTLE     = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	ORANGE     = lipgloss.AdaptiveColor{Light: "214", Dark: "214"}
	DEFAULT_FG = lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
)

func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}
