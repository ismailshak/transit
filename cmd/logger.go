package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const black = "#000000"

var (
	errorPrefix = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FF2A00")).Foreground(lipgloss.Color(black)).Render("Error")
	warnPrefix  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FFC34D")).Foreground(lipgloss.Color(black)).Render("Warn")
)

// errorf writes to Err.
func (a *App) errorf(format string, args ...any) {
	_, _ = fmt.Fprintln(a.Err, errorPrefix, fmt.Sprintf(format, args...))
}

// warnf writes to Err.
func (a *App) warnf(format string, args ...any) {
	_, _ = fmt.Fprintln(a.Err, warnPrefix, fmt.Sprintf(format, args...))
}
