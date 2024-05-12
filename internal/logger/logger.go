// Package logger implements functions that print messages to the user's terminal.
//
// Follows a standard server-logger approach, which contains functions for writing errors to stderr
// and messages to stdout. This package should not import other transit packages (excluding config)
// to avoid import cycles. This package is not used to display pretty terminal renderings,
// that is handled by the tui package
package logger

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/config"
)

const BLACK = "#000000"

var (
	DEBUG_PREFIX = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#67CBFF")).Foreground(lipgloss.Color(BLACK)).Render("Debug")
	ERR_PREFIX   = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FF2A00")).Foreground(lipgloss.Color(BLACK)).Render("Error")
	INFO_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#B3E5FF")).Foreground(lipgloss.Color(BLACK)).Render("Info")
	WARN_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FFC34D")).Foreground(lipgloss.Color(BLACK)).Render("Warn")
)

func Debug(message ...any) {
	if config.GetConfig().Core.Verbose {
		fmt.Println(DEBUG_PREFIX, fmt.Sprint(message...))
	}
}

func Error(message ...any) {
	fmt.Println(ERR_PREFIX, fmt.Sprint(message...))
}

func Info(message ...any) {
	fmt.Println(INFO_PREFIX, fmt.Sprint(message...))
}

func Warn(message ...any) {
	fmt.Println(WARN_PREFIX, fmt.Sprint(message...))
}

// Print to standard out without any formatting or prefixes
func Print(message ...any) {
	fmt.Println(message...)
}
