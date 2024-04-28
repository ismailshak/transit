package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
	"golang.org/x/term"
)

const (
	DATE_FORMAT = "2 Jan 06 3:04pm"
)

func PrintIncidents(client api.Api, incidents []api.Incident) {
	if len(incidents) == 0 {
		logger.Print("No incidents reported")
		return
	}

	maxWidth := 80
	termWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
	width := termWidth - 5 // some padding

	if width > maxWidth {
		width = maxWidth
	}

	for _, inc := range incidents {
		render(client, inc, width)
	}
}

func formatUpdatedAt(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return date.Format(DATE_FORMAT)
}

func formatStartEnd(start, end time.Time) string {
	if start.IsZero() && end.IsZero() {
		return ""
	}

	if start.IsZero() {
		return fmt.Sprintf("Ends: %s", end.Format(DATE_FORMAT))
	}

	if end.IsZero() {
		return fmt.Sprintf("Starts: %s", end.Format(DATE_FORMAT))
	}

	return fmt.Sprintf("%s - %s", start.Format(DATE_FORMAT), end.Format(DATE_FORMAT))
}

func genFooter(incident *api.Incident) string {
	duration := formatStartEnd(incident.ActivePeriodStart, incident.ActivePeriodEnd)
	agencyName := incident.Agency

	if agencyName == "" && duration == "" {
		return ""
	}

	activePeriod := utils.Ternary(duration != "", lipgloss.NewStyle().Margin(1, 1, 0).Render(duration), "")

	agencyHorMargin := utils.Ternary(duration == "", 1, 2) // If duration will render, add extra margin
	agency := utils.Ternary(incident.Agency != "", lipgloss.NewStyle().Margin(1, agencyHorMargin, 0).Foreground(lipgloss.Color("30")).Render(incident.Agency), "")

	return lipgloss.JoinHorizontal(lipgloss.Left, activePeriod, agency)
}

func render(client api.Api, incident api.Incident, width int) {
	list := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		Padding(1, 1).
		BorderForeground(SUBTLE)

	incType := lipgloss.NewStyle().Padding(0, 1).Bold(true).Render(incident.Type)

	update := lipgloss.NewStyle().Margin(0, 1).Render(formatUpdatedAt(incident.DateUpdated))

	affected := genAffected(client, incident.Affected)

	header := lipgloss.JoinHorizontal(lipgloss.Left, incType, affected, update)

	description := lipgloss.NewStyle().Width(width).Margin(1, 1, 0).Render(incident.Description)

	footer := genFooter(&incident)

	// TODO Clean up UI
	if footer == "" {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description))
		logger.Print(out)
	} else {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description, footer))
		logger.Print(out)
	}
}

func genAffected(client api.Api, affected []string) string {
	builder := strings.Builder{}

	for _, a := range affected {
		bg, fg := client.GetLineColor(a)
		line := lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg)).Render(a)
		builder.WriteString(line)
	}

	return builder.String()
}
