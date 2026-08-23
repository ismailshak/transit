package render_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/render"
	"github.com/ismailshak/transit/internal/transit"
)

func TestAlerts(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		alerts     []transit.Alert
		width      int
		showAgency bool
	}{
		"no_alerts": {
			alerts:     nil,
			width:      testWidth,
			showAgency: true,
		},
		"agency_ref": {
			alerts:     []transit.Alert{agencyAlert()},
			width:      testWidth,
			showAgency: true,
		},
		"route_refs": {
			alerts:     []transit.Alert{routeAlert()},
			width:      testWidth,
			showAgency: true,
		},
		"stop_refs": {
			alerts:     []transit.Alert{stopsAlert()},
			width:      testWidth,
			showAgency: false,
		},
		"stop_refs_narrow": {
			alerts:     []transit.Alert{stopsAlert()},
			width:      50,
			showAgency: false,
		},
		"no_timestamp": {
			alerts:     []transit.Alert{undatedAlert()},
			width:      testWidth,
			showAgency: true,
		},
		"two_alerts": {
			alerts:     []transit.Alert{agencyAlert(), stopsAlert()},
			width:      testWidth,
			showAgency: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			alerts := render.Alerts{
				Set:        transit.AlertSet{Alerts: tc.alerts},
				Width:      tc.width,
				ShowAgency: tc.showAgency,
			}

			golden(t, name, alerts.String())
		})
	}
}
