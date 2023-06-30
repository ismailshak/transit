package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/logger"
	"golang.org/x/term"
)

const (
	DATE_FORMAT = "2 Jan 06 3:04pm"
)

func PrintIssues(client api.Api, incidents []api.Incident) {
	if len(incidents) == 0 {
		logger.Print("No incidents reported")
		return
	}

	maxWidth := 80
	termWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
	width := termWidth - 10 // some padding

	if width > maxWidth {
		width = maxWidth
	}

	for _, inc := range incidents {
		render(client, inc, width)
	}
}

func render(client api.Api, incident api.Incident, width int) {
	list := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		Padding(1, 1).
		BorderForeground(subtle)

	inc_type := lipgloss.NewStyle().Padding(0, 1).Bold(true).Render(incident.Type)

	update := lipgloss.NewStyle().Margin(0, 1).Render(incident.DateUpdated.Format(DATE_FORMAT))

	affected := genAffected(client, incident.Affected)

	header := lipgloss.JoinHorizontal(lipgloss.Left, inc_type, affected, update)

	description := lipgloss.NewStyle().Width(width).Margin(1, 1, 0).Render(incident.Description)

	out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description))
	logger.Print(out)
}

func genAffected(client api.Api, affected []string) string {
	builder := strings.Builder{}

	for _, a := range affected {
		bg, fg := client.GetColorFromLine(a)
		line := lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg)).Render(a)
		builder.WriteString(line)
	}

	return builder.String()
}
