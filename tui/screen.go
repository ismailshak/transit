package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
)

var (
	subtle     = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	orange     = lipgloss.AdaptiveColor{Light: "214", Dark: "214"}
	default_fg = lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
)

// Create and return a terminal layout that will contain the screen-like display
func getScreen() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, false).
		BorderForeground(subtle)
}

// Generate the header that will be printed at the top of the screen
func genHeader(header string) string {
	return lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(subtle).
		PaddingTop(1).
		Render(header)
}

// Generates a row printed on the screen
func genRow(line, destination string, minutes []string) string {
	formattedLine := genLine(line)
	formattedDest := genDestination(destination)
	formattedMins := genTimeList(minutes)

	return lipgloss.JoinHorizontal(lipgloss.Left, formattedLine, formattedDest, formattedMins)
}

// Generate and color a metro's line
func genLine(line string) string {
	bg, fg := helpers.GetColorFromLine(line)
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
func genTimeList(minutes []string) string {
	formatted := []string{}
	for _, m := range minutes {
		formatted = append(formatted, genTimeEntry(m))
	}

	return strings.Join(formatted, ",")
}

// Generate a formatted entry for a single ETA
func genTimeEntry(time string) string {
	return lipgloss.NewStyle().
		Foreground(orange).
		Align(lipgloss.Right).
		Render(time)
}

// Create and print a screen that resembles a station's. Will display
// an arriving train's line, destination and arriving trains (in "minutes-away")
func PrintArrivingScreen(predictions []api.Predictions) {
	list := getScreen()
	header := predictions[0].LocationName

	destinationLookup := groupByDestination(predictions)
	items := []string{}
	items = append(items, genHeader(header))
	for _, v := range destinationLookup {
		if helpers.IsGhostTrain(v.line, v.destination) {
			logger.Debug(("A train not intended for passengers is hidden from the display"))
			logger.Debug(fmt.Sprintf("%+v", v))
			continue
		}

		item := genRow(v.line, v.destination, v.minutes)
		items = append(items, item)
	}

	out := list.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			items...,
		),
	)

	builder := strings.Builder{}
	builder.WriteString(out)

	logger.Print(builder.String())
}

type Row struct {
	destination string
	line        string
	minutes     []string
}

func groupByDestination(predictions []api.Predictions) map[string]*Row {
	destMap := make(map[string]*Row)
	for _, t := range predictions {
		_, exists := destMap[t.Destination]
		if exists {
			destMap[t.Destination].minutes = append(destMap[t.Destination].minutes, t.Min)
		} else {
			destMap[t.Destination] = &Row{
				destination: t.Destination,
				line:        t.Line,
				minutes:     []string{t.Min},
			}

		}
	}

	return destMap
}
