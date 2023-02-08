package tui

import "github.com/charmbracelet/lipgloss"

func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}
