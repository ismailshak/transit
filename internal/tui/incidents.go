package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/provider"
	"golang.org/x/term"
)

const (
	dateFormat = "2 Jan 06 3:04pm"
)

func PrintIncidents(client provider.API, incidents []provider.Incident) {
	if len(incidents) == 0 {
		fmt.Println("No incidents reported")
		return
	}

	maxWidth := 80
	termWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
	width := min(max(termWidth-5, 0), maxWidth) // -5 for some padding

	for _, inc := range incidents {
		render(client, inc, width)
	}
}

func formatUpdatedAt(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return date.Format(dateFormat)
}

func formatStartEnd(start, end time.Time) string {
	if start.IsZero() && end.IsZero() {
		return ""
	}

	if start.IsZero() {
		return fmt.Sprintf("Ends: %s", end.Format(dateFormat))
	}

	if end.IsZero() {
		return fmt.Sprintf("Starts: %s", end.Format(dateFormat))
	}

	return fmt.Sprintf("%s - %s", start.Format(dateFormat), end.Format(dateFormat))
}

func genFooter(incident *provider.Incident) string {
	duration := formatStartEnd(incident.ActivePeriodStart, incident.ActivePeriodEnd)
	agencyName := incident.Agency

	if agencyName == "" && duration == "" {
		return ""
	}

	var activePeriod string
	if duration != "" {
		activePeriod = lipgloss.NewStyle().Margin(1, 1, 0).Render(duration)
	}

	var agencyHorMargin int
	if duration == "" {
		agencyHorMargin = 1
	} else {
		agencyHorMargin = 2
	}

	var agency string
	if incident.Agency != "" {
		agency = lipgloss.NewStyle().Margin(1, agencyHorMargin, 0).Foreground(lipgloss.Color("30")).Render(incident.Agency)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, activePeriod, agency)
}

func render(client provider.API, incident provider.Incident, width int) {
	list := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		Padding(1, 1).
		BorderForeground(Subtle)

	incType := lipgloss.NewStyle().Padding(0, 1).Bold(true).Render(incident.Type)

	update := lipgloss.NewStyle().Margin(0, 1).Render(formatUpdatedAt(incident.DateUpdated))

	affected := genAffected(client, incident.Affected)

	header := lipgloss.JoinHorizontal(lipgloss.Left, incType, affected, update)

	description := lipgloss.NewStyle().Width(width).Margin(1, 1, 0).Render(incident.Description)

	footer := genFooter(&incident)

	// TODO Clean up UI
	if footer == "" {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description))
		fmt.Println(out)
	} else {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description, footer))
		fmt.Println(out)
	}
}

func genAffected(client provider.API, affected []string) string {
	builder := strings.Builder{}

	for _, a := range affected {
		bg, fg := client.GetLineColor(a)
		line := lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg)).Render(a)
		builder.WriteString(line)
	}

	return builder.String()
}
