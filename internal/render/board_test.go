package render_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/render"
	"github.com/ismailshak/transit/internal/transit"
)

func TestBoard(t *testing.T) {
	t.Parallel()

	tests := map[string][]transit.Departure{
		"one_station":   metroCenter(),
		"two_stations":  append(metroCenter(), courthouse()...),
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
