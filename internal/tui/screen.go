package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/transit"
)

// PrintArrivalScreen creates and prints a screen that resembles a station's. Will display
// an arriving train's line, destination and arriving trains (in "minutes-away").
func PrintArrivalScreen(destinationLookup *map[string][]transit.Departure, sortedDestinations []string, now time.Time) {
	list := getScreen()

	// since this is the same for all items, fishing it out from the first one
	header := (*destinationLookup)[sortedDestinations[0]][0].StopName

	items := []string{}
	items = append(items, genHeader(header))

	for _, d := range sortedDestinations {
		items = append(items, genRow((*destinationLookup)[d], now))
	}

	out := list.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			items...,
		),
	)

	fmt.Println(out)
}

// Create and return a terminal layout that will contain the screen-like display
func getScreen() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, false).
		BorderForeground(Subtle)
}

// Generate the header that will be printed at the top of the screen
func genHeader(header string) string {
	return lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Subtle).
		PaddingTop(1).
		Render(header)
}

// Generates a row printed on the screen
func genRow(destination []transit.Departure, now time.Time) string {
	formattedLine := genLine(destination[0])
	formattedDest := genDestination(destination[0].Headsign)
	formattedMins := genTimeList(destination, now)

	return lipgloss.JoinHorizontal(lipgloss.Left, formattedLine, formattedDest, formattedMins)
}

// Generate and color a metro's line
func genLine(d transit.Departure) string {
	return lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(d.LineColor)).
		Foreground(lipgloss.Color(d.LineText)).
		Padding(0, 1).
		Render(d.Line)
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
func genTimeList(destination []transit.Departure, now time.Time) string {
	formatted := []string{}
	for _, d := range destination {
		formatted = append(formatted, genTimeEntry(d.Arrives, now))
	}

	return strings.Join(formatted, ",")
}

// Generate a formatted entry for a single ETA
func genTimeEntry(arrives, now time.Time) string {
	return lipgloss.NewStyle().
		Foreground(Orange).
		Align(lipgloss.Right).
		Render(minutesAway(arrives, now))
}

func minutesAway(arrives, now time.Time) string {
	mins := int(arrives.Sub(now).Round(time.Minute) / time.Minute)
	return strconv.Itoa(max(mins, 0))
}
