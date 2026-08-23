package render

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/transit"
)

const (
	dateFormat               = "2 Jan 06 3:04pm"
	maxAlertDescriptionWidth = 72
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			Padding(1, 1).
			BorderForeground(Subtle)

	effectStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	affectedStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 1)

	descriptionStyle = lipgloss.NewStyle().
				Margin(1, 1, 0)

	asOfStyle = lipgloss.NewStyle().
			Margin(0, 1).
			Faint(true)

	activePeriodStyle = lipgloss.NewStyle().
				Margin(1, 1, 0)

	agencyStyle = lipgloss.NewStyle().
			Foreground(Cyan)
)

// Alerts renders a notice box for each alert in the set.
type Alerts struct {
	Set        transit.AlertSet
	Width      int // The terminal's width. Will use internal fallback when zero.
	ShowAgency bool
}

// String renders all alerts, one box per alert.
func (a Alerts) String() string {
	if len(a.Set.Alerts) == 0 {
		return "No alerts reported"
	}

	descriptionWidth := maxAlertDescriptionWidth
	if a.Width > 0 {
		descriptionWidth = min(
			a.Width-boxStyle.GetHorizontalFrameSize()-descriptionStyle.GetHorizontalMargins(),
			maxAlertDescriptionWidth,
		)
	}

	items := make([]string, 0, len(a.Set.Alerts))
	for _, alert := range a.Set.Alerts {
		items = append(items, box(alert, descriptionWidth, a.ShowAgency))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func box(alert transit.Alert, width int, showAgency bool) string {
	blocks := []string{head(alert), descriptionStyle.Width(width).Render(alert.Description)}

	if f := footer(alert, showAgency); f != "" {
		blocks = append(blocks, f)
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, blocks...))
}

func head(alert transit.Alert) string {
	blocks := []string{effectStyle.Render(alert.Effect)}

	for _, a := range alert.Affected {
		blocks = append(blocks, ref(a))
	}

	if asOf := formatUpdatedAt(alert.Updated); asOf != "" {
		blocks = append(blocks, asOfStyle.Render(asOf))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}

// ref is one entity's badge. Only the branded ones carry the agency's colors.
func ref(a transit.AlertRef) string {
	if a.Kind == transit.RefRoute {
		return affectedStyle.
			Background(lipgloss.Color(a.Color)).
			Foreground(lipgloss.Color(a.TextColor)).
			Render(a.ID)
	}

	return affectedStyle.
		Border(lipgloss.NormalBorder(), true, true).
		BorderForeground(Subtle).
		Foreground(Subtle).
		Render(a.ID)
}

func footer(alert transit.Alert, showAgency bool) string {
	duration := formatStartEnd(alert)

	var agencyID string
	if showAgency {
		agencyID = alert.AgencyID
	}

	if agencyID == "" && duration == "" {
		return ""
	}

	var activePeriod string
	if duration != "" {
		activePeriod = activePeriodStyle.Render(duration)
	}

	agencyHorMargin := 1
	if duration != "" {
		agencyHorMargin = 2
	}

	var agency string
	if agencyID != "" {
		agency = agencyStyle.Margin(1, agencyHorMargin, 0).Render(agencyID)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, activePeriod, agency)
}

func formatUpdatedAt(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return "as of " + date.Format(dateFormat)
}

func formatStartEnd(alert transit.Alert) string {
	if alert.Starts.IsZero() && alert.Ends.IsZero() {
		return ""
	}

	if alert.Starts.IsZero() {
		return fmt.Sprintf("Ends: %s", alert.Ends.Format(dateFormat))
	}

	if alert.Ends.IsZero() {
		return fmt.Sprintf("Starts: %s", alert.Starts.Format(dateFormat))
	}

	return fmt.Sprintf("%s - %s", alert.Starts.Format(dateFormat), alert.Ends.Format(dateFormat))
}
