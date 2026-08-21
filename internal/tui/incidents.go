package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/transit"
	"golang.org/x/term"
)

const (
	dateFormat = "2 Jan 06 3:04pm"
)

func PrintIncidents(client provider.API, alertSet transit.AlertSet, showAgency bool) {
	if len(alertSet.Alerts) == 0 {
		fmt.Println("No incidents reported")
		return
	}

	maxWidth := 80
	termWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
	width := min(max(termWidth-5, 0), maxWidth) // -5 for some padding

	for _, a := range alertSet.Alerts {
		render(client, a, width, showAgency)
	}

	// TODO: Print once
	fmt.Println(lipgloss.NewStyle().Margin(1, 1).Faint(true).Render(formatUpdatedAt(alertSet.AsOf())))
}

func formatUpdatedAt(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return "as of " + date.Format(dateFormat)
}

func formatStartEnd(start, end time.Time) string {
	if start.IsZero() && end.IsZero() {
		return ""
	}

	if start.IsZero() {
		return fmt.Sprintf("Ends: %s", end.Format(dateFormat))
	}

	if end.IsZero() {
		return fmt.Sprintf("Starts: %s", start.Format(dateFormat))
	}

	return fmt.Sprintf("%s - %s", start.Format(dateFormat), end.Format(dateFormat))
}

func genFooter(alert *transit.Alert, showAgency bool) string {
	duration := formatStartEnd(alert.Starts, alert.Ends)

	var agencyID string
	if showAgency {
		agencyID = alert.AgencyID
	}

	if agencyID == "" && duration == "" {
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
	if agencyID != "" {
		agency = lipgloss.NewStyle().Margin(1, agencyHorMargin, 0).Foreground(lipgloss.Color("30")).Render(agencyID)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, activePeriod, agency)
}

func render(client provider.API, alert transit.Alert, width int, showAgency bool) {
	list := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, true, true, true).
		Padding(1, 1).
		BorderForeground(Subtle)

	effect := lipgloss.NewStyle().Padding(0, 1).Bold(true).Render(alert.Effect)

	affected := genAffected(client, alert.Affected)

	header := lipgloss.JoinHorizontal(lipgloss.Left, effect, affected)

	description := lipgloss.NewStyle().Width(width).Margin(1, 1, 0).Render(alert.Description)

	footer := genFooter(&alert, showAgency)

	// TODO Clean up UI
	if footer == "" {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description))
		fmt.Println(out)
	} else {
		out := list.Render(lipgloss.JoinVertical(lipgloss.Left, header, description, footer))
		fmt.Println(out)
	}
}

func genAffected(client provider.API, affected []transit.AlertRef) string {
	builder := strings.Builder{}

	for _, a := range affected {
		style := lipgloss.NewStyle().Padding(0, 1).Margin(0, 1)

		if a.Kind == transit.RefRoute {
			bg, fg := client.GetLineColor(a.ID)
			style = style.Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(fg))
		} else {
			style = style.Border(lipgloss.NormalBorder(), true, true).BorderForeground(Subtle).Foreground(Subtle)
		}

		builder.WriteString(style.Render(a.ID))
	}

	return builder.String()
}
