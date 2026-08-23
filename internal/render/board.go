package render

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismailshak/transit/internal/transit"
)

var (
	boardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, false).
			BorderForeground(Subtle)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(Subtle).
			PaddingTop(1)

	lineStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	destinationStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				PaddingRight(3).
				PaddingBottom(1).
				Width(20)

	etaStyle = lipgloss.NewStyle().
			Foreground(Orange).
			Align(lipgloss.Right)
)

// Board is a station's arrival screen.
type Board struct {
	Set transit.DepartureSet
	Now time.Time
}

// String renders the screen, one section per stop.
func (b Board) String() string {
	stops := groupByStop(b.Set.Departures)

	var items []string
	for _, s := range stops {
		items = append(items, header(s.name))
		for _, r := range s.rows {
			items = append(items, row(r, b.Now))
		}
	}

	return boardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, items...))
}

// stopGroup is one stop's section. A row is the departures sharing a line and headsign.
type stopGroup struct {
	name string
	rows [][]transit.Departure
}

// Stops stay in the order their departures arrived in. Ranging a map would shuffle them.
func groupByStop(departures []transit.Departure) []stopGroup {
	var names []string
	byStop := make(map[string][]transit.Departure)

	for _, d := range departures {
		if _, ok := byStop[d.StopName]; !ok {
			names = append(names, d.StopName)
		}

		byStop[d.StopName] = append(byStop[d.StopName], d)
	}

	stops := make([]stopGroup, 0, len(names))
	for _, n := range names {
		stops = append(stops, stopGroup{name: n, rows: groupByDestination(byStop[n])})
	}

	return stops
}

// The same destination is sometimes served by more than one line so we have to key by
// a combination of both.
func groupByDestination(departures []transit.Departure) [][]transit.Departure {
	var keys []string
	byDest := make(map[string][]transit.Departure)

	for _, d := range departures {
		key := d.Headsign + "-" + d.Line
		if _, ok := byDest[key]; !ok {
			keys = append(keys, key)
		}

		byDest[key] = append(byDest[key], d)
	}

	slices.Sort(keys)

	rows := make([][]transit.Departure, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, byDest[k])
	}

	return rows
}

func header(stop string) string {
	return headerStyle.Render(stop)
}

func row(departures []transit.Departure, now time.Time) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		line(departures[0]),
		destination(departures[0].Headsign),
		etas(departures, now),
	)
}

func line(d transit.Departure) string {
	return lineStyle.
		Background(lipgloss.Color(d.LineColor)).
		Foreground(lipgloss.Color(d.LineText)).
		Render(d.Line)
}

func destination(headsign string) string {
	return destinationStyle.Render(headsign)
}

func etas(departures []transit.Departure, now time.Time) string {
	formatted := make([]string, 0, len(departures))
	for _, d := range departures {
		formatted = append(formatted, eta(d.Arrives, now))
	}

	return strings.Join(formatted, ",")
}

func eta(arrives, now time.Time) string {
	return etaStyle.Render(minutesAway(arrives, now))
}

func minutesAway(arrives, now time.Time) string {
	mins := int(arrives.Sub(now).Round(time.Minute) / time.Minute)
	return strconv.Itoa(max(mins, 0))
}
