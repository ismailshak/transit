package logger

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/config"
)

const BLACK = "#000000"

var (
	DEBUG_PREFIX = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#67CBFF")).Foreground(lipgloss.Color(BLACK)).Render("Debug")
	ERR_PREFIX   = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FF2A00")).Foreground(lipgloss.Color(BLACK)).Render("Error")
	INFO_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#B3E5FF")).Foreground(lipgloss.Color(BLACK)).Render("Info")
	WARN_PREFIX  = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#FFC34D")).Foreground(lipgloss.Color(BLACK)).Render("Warn")
)

func Debug(message string) {
	if config.GetConfig().Core.Verbose {
		fmt.Println(DEBUG_PREFIX, message)
	}
}

func Error(message string) {
	fmt.Println(ERR_PREFIX, message)
}

func Info(message string) {
	fmt.Println(INFO_PREFIX, message)
}

func Warn(message string) {
	fmt.Println(WARN_PREFIX, message)
}

// Print to standard out without any formatting or prefixes
func Print(message ...any) {
	fmt.Println(message...)
}
