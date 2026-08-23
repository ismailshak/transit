package render_test

import (
	"time"

	"github.com/ismailshak/transit/internal/transit"
)

var now = time.Date(2026, time.August, 22, 9, 30, 0, 0, time.UTC)

// The terminal width the goldens are recorded at.
const testWidth = 80

const (
	redBg    = "#BF0D3E"
	blueBg   = "#009CDE"
	orangeBg = "#ED8B00"
	silverBg = "#919D9D"
	yellowBg = "#FFD100"
	greenBg  = "#00B140"

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
func metroCenter() []transit.Departure {
	return []transit.Departure{
		departure("Metro Center", "RD", redBg, onDark, "Shady Grove", 3*time.Minute),
		departure("Metro Center", "RD", redBg, onDark, "Glenmont", 5*time.Minute),
		departure("Metro Center", "BL", blueBg, onDark, "Largo", 7*time.Minute),
		departure("Metro Center", "SV", silverBg, onLight, "Largo", 2*time.Minute),
		departure("Metro Center", "RD", redBg, onDark, "Shady Grove", 9*time.Minute),
		departure("Metro Center", "OR", orangeBg, onLight, "New Carrollton", -1*time.Minute),
	}
}

func courthouse() []transit.Departure {
	return []transit.Departure{
		departure("Courthouse", "SV", silverBg, onLight, "Ashburn", 4*time.Minute),
		departure("Courthouse", "OR", orangeBg, onLight, "Vienna", 6*time.Minute),
		departure("Courthouse", "SV", silverBg, onLight, "Ashburn", 14*time.Minute),
	}
}

func stopRefs(ids ...string) []transit.AlertRef {
	refs := make([]transit.AlertRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, transit.AlertRef{Kind: transit.RefStop, ID: id})
	}

	return refs
}

// Just a single agency affected. No start and end so no footer.
func agencyAlert() transit.Alert {
	return transit.Alert{
		Source:      "bayarea511",
		AgencyID:    "BA",
		Affected:    []transit.AlertRef{{Kind: transit.RefAgency, ID: "BA"}},
		Effect:      "Notice",
		Description: "BART's schedule will change on Monday, August 24. Some train times are adjusting by several minutes. Use the BART Trip Planner with a date of August 24 or after to see if your trip has changed.",
		Updated:     now.Add(-90 * time.Minute),
	}
}

// The same agencyAlert fixture but with Updated zeroed out.
func undatedAlert() transit.Alert {
	a := agencyAlert()
	a.Updated = time.Time{}

	return a
}

// Branded colors since these are routes affected.
func routeAlert() transit.Alert {
	return transit.Alert{
		Source:   "wmata-rail",
		AgencyID: "MET",
		Affected: []transit.AlertRef{
			{Kind: transit.RefRoute, ID: "YL", Color: yellowBg, TextColor: onLight},
			{Kind: transit.RefRoute, ID: "GR", Color: greenBg, TextColor: onDark},
		},
		Effect:      "Alert",
		Description: "Yellow Line trains are operating between Huntington and Mt Vernon Square only as a result of a track problem outside College Park. Customers can continue to Greenbelt by boarding a Green Line train.",
		Updated:     now.Add(-30 * time.Minute),
	}
}

// Enough stops to run the header past the box. The active period gives it a footer.
func stopsAlert() transit.Alert {
	return transit.Alert{
		Source:      "bayarea511",
		AgencyID:    "BA",
		Affected:    stopRefs("EMBR", "MONT", "POWL", "CIVC", "16TH", "24TH", "GLEN", "BALB", "DALY", "COLM"),
		Effect:      "Significant Delays",
		Description: "Trains are single tracking between Embarcadero and Colma while crews replace rail. Expect delays of up to 20 minutes in both directions. Free bus service runs between Daly City and Colma.",
		Starts:      time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC),
		Ends:        time.Date(2026, time.August, 24, 4, 0, 0, 0, time.UTC),
		Updated:     now.Add(-3 * time.Hour),
	}
}
