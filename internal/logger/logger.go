// Package logger implements functions that print messages to the user's terminal.
//
// Follows a standard server-logger approach, which contains functions for writing errors to stderr
// and messages to stdout. This package should not import other transit packages
// to avoid import cycles. This package is not used to display pretty terminal renderings,
// that is handled by the tui package
package logger

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const BLACK = "#000000"

var (
	DEBUG_PREFIX = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#67CBFF")).Foreground(lipgloss.Color(BLACK)).Render("Debug")
	ERR_PREFIX   = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FF2A00")).Foreground(lipgloss.Color(BLACK)).Render("Error")
	INFO_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#B3E5FF")).Foreground(lipgloss.Color(BLACK)).Render("Info")
	WARN_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FFC34D")).Foreground(lipgloss.Color(BLACK)).Render("Warn")
)

// Logger writes diagnostics for one run of the program. Debug messages are
// dropped unless verbose logging was asked for.
type Logger struct {
	verbose bool
}

// New returns a Logger. Pass verbose to let Debug messages through.
func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Debug writes a message, but only when verbose logging is on
func (l *Logger) Debug(message ...any) {
	if l.verbose {
		fmt.Println(DEBUG_PREFIX, fmt.Sprint(message...))
	}
}

func (l *Logger) Error(message ...any) {
	fmt.Println(ERR_PREFIX, fmt.Sprint(message...))
}

func (l *Logger) Info(message ...any) {
	fmt.Println(INFO_PREFIX, fmt.Sprint(message...))
}

func (l *Logger) Warn(message ...any) {
	fmt.Println(WARN_PREFIX, fmt.Sprint(message...))
}

// Verbose reports whether Debug will write anything. Call it to skip building a
// message that would only be thrown away.
func (l *Logger) Verbose() bool {
	return l.verbose
}
