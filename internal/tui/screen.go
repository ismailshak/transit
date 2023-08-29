package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/pkg/api"
)

// Create and print a screen that resembles a station's. Will display
// an arriving train's line, destination and arriving trains (in "minutes-away")
func PrintArrivingScreen(client api.Api, destinationLookup *map[string][]api.Prediction, sortedDestinations []string) {
	list := getScreen()

	// since this is the same for all items, fishing it out from the first one
	header := (*destinationLookup)[sortedDestinations[0]][0].LocationName

	items := []string{}
	items = append(items, genHeader(header))

	for _, d := range sortedDestinations {
		destination := (*destinationLookup)[d]
		if client.IsGhostTrain(destination[0].Line, destination[0].Destination) {
			logger.Debug(("A train not intended for passengers is hidden from the display"))
			logger.Debug(fmt.Sprintf("%+v", destination[0]))
			continue
		}

		item := genRow(client, destination)
		items = append(items, item)
	}

	out := list.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			items...,
		),
	)

	logger.Print(out)
}

// Create and return a terminal layout that will contain the screen-like display
func getScreen() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, false).
		BorderForeground(SUBTLE)
}

// Generate the header that will be printed at the top of the screen
func genHeader(header string) string {
	return lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(SUBTLE).
		PaddingTop(1).
		Render(header)
}

// Generates a row printed on the screen
func genRow(client api.Api, destination []api.Prediction) string {
	formattedLine := genLine(client, destination[0].Line)
	formattedDest := genDestination(destination[0].Destination)
	formattedMins := genTimeList(destination)

	return lipgloss.JoinHorizontal(lipgloss.Left, formattedLine, formattedDest, formattedMins)
}

// Generate and color a metro's line
func genLine(client api.Api, line string) string {
	bg, fg := client.GetStopColor(line)
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Padding(0, 1).
		Render(line)
}

// Generate a formatted (and padded) destination item
func genDestination(destination string) string {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(3).
		PaddingBottom(1).
		Width(20).
		Render(destination)
}

// Generates a comma separated list of formatted minutes until
func genTimeList(destination []api.Prediction) string {
	formatted := []string{}
	for _, d := range destination {
		formatted = append(formatted, genTimeEntry(d.Min))
	}

	return strings.Join(formatted, ",")
}

// Generate a formatted entry for a single ETA
func genTimeEntry(time string) string {
	return lipgloss.NewStyle().
		Foreground(ORANGE).
		Align(lipgloss.Right).
		Render(time)
}
