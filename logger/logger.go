package logger

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/config"
)

var (
	ERR_PREFIX   = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FF2A00")).Foreground(lipgloss.Color("#000000")).Render("Error")
	WARN_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FFC34D")).Foreground(lipgloss.Color("#000000")).Render("Warn")
	DEBUG_PREFIX = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#B3E5FF")).Foreground(lipgloss.Color("#000000")).Render("Debug")
)

func Error(message string) {
	fmt.Println(ERR_PREFIX + " " + message)
}

func Warn(message string) {
	fmt.Println(WARN_PREFIX + " " + message)
}

func Debug(message string) {
	if config.GetConfig().Core.Verbose {
		fmt.Println(DEBUG_PREFIX + " " + message)
	}
}

// Print to standard out without any formatting or prefixes
func Print(message string) {
	fmt.Println(message)
}
