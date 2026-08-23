package render_test

import (
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/render"
	"github.com/ismailshak/transit/internal/transit"
)

var now = time.Date(2026, time.August, 22, 9, 30, 0, 0, time.UTC)

const (
	redBg    = "#BF0D3E"
	blueBg   = "#009CDE"
	orangeBg = "#ED8B00"
	silverBg = "#919D9D"

	onDark  = "#FFFFFF"
	onLight = "#000000"
)

func departure(stop, line, bg, text, headsign string, in time.Duration) transit.Departure {
	return transit.Departure{
		Source:    "wmata",
		StopName:  stop,
		AgencyID:  "MET",
		Mode:      transit.ModeMetro,
		Line:      line,
		LineColor: bg,
		LineText:  text,
		Headsign:  headsign,
		Arrives:   now.Add(in),
	}
}

// metroCenter covers: several trains on one row, one destination served
// by two lines, and a train that has already left the station.
var metroCenter = []transit.Departure{
	departure("Metro Center", "RD", redBg, onDark, "Shady Grove", 3*time.Minute),
	departure("Metro Center", "RD", redBg, onDark, "Glenmont", 5*time.Minute),
	departure("Metro Center", "BL", blueBg, onDark, "Largo", 7*time.Minute),
	departure("Metro Center", "SV", silverBg, onLight, "Largo", 2*time.Minute),
	departure("Metro Center", "RD", redBg, onDark, "Shady Grove", 9*time.Minute),
	departure("Metro Center", "OR", orangeBg, onLight, "New Carrollton", -1*time.Minute),
}

var courthouse = []transit.Departure{
	departure("Courthouse", "SV", silverBg, onLight, "Ashburn", 4*time.Minute),
	departure("Courthouse", "OR", orangeBg, onLight, "Vienna", 6*time.Minute),
	departure("Courthouse", "SV", silverBg, onLight, "Ashburn", 14*time.Minute),
}

func TestBoard(t *testing.T) {
	t.Parallel()

	tests := map[string][]transit.Departure{
		"one_station":   metroCenter,
		"two_stations":  append(append([]transit.Departure{}, metroCenter...), courthouse...),
		"no_departures": nil,
	}

	for name, departures := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			board := render.Board{
				Set: transit.DepartureSet{Departures: departures},
				Now: now,
			}

			golden(t, name, board.String())
		})
	}
}
