package render

import (
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/transit"
)

func TestFormatUpdatedAt(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		date time.Time
		want string
	}{
		"zero date": {
			date: time.Time{},
			want: "",
		},
		"morning, single-digit day": {
			date: time.Date(2023, time.June, 1, 7, 27, 0, 0, time.UTC),
			want: "as of 1 Jun 23 7:27am",
		},
		"afternoon, double-digit day": {
			date: time.Date(2020, time.October, 11, 9, 1, 0, 0, time.UTC),
			want: "as of 11 Oct 20 9:01am",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := formatUpdatedAt(tt.date)

			if got != tt.want {
				t.Errorf("expected %q but got %q", tt.want, got)
			}
		})
	}
}

func TestFormatStartEnd(t *testing.T) {
	t.Parallel()

	starts := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	ends := time.Date(2026, time.August, 24, 16, 30, 0, 0, time.UTC)

	tests := map[string]struct {
		alert transit.Alert
		want  string
	}{
		"neither": {
			alert: transit.Alert{},
			want:  "",
		},
		"start only": {
			alert: transit.Alert{Starts: starts},
			want:  "Starts: 22 Aug 26 12:00am",
		},
		"end only": {
			alert: transit.Alert{Ends: ends},
			want:  "Ends: 24 Aug 26 4:30pm",
		},
		"both": {
			alert: transit.Alert{Starts: starts, Ends: ends},
			want:  "22 Aug 26 12:00am - 24 Aug 26 4:30pm",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := formatStartEnd(tt.alert)

			if got != tt.want {
				t.Errorf("expected %q but got %q", tt.want, got)
			}
		})
	}
}
