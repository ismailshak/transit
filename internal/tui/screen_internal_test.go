package tui

import (
	"testing"
	"time"
)

func TestMinutesAway(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 21, 9, 30, 0, 0, time.UTC)

	tests := map[string]struct {
		arrives time.Time
		want    string
	}{
		"arriving now": {
			arrives: now,
			want:    "0",
		},
		"whole minutes away": {
			arrives: now.Add(5 * time.Minute),
			want:    "5",
		},
		"rounds down below the half minute": {
			arrives: now.Add(4*time.Minute + 29*time.Second),
			want:    "4",
		},
		"rounds up on the half minute": {
			arrives: now.Add(4*time.Minute + 30*time.Second),
			want:    "5",
		},
		"under a minute away": {
			arrives: now.Add(20 * time.Second),
			want:    "0",
		},
		"more than an hour away": {
			arrives: now.Add(95 * time.Minute),
			want:    "95",
		},
		"already departed": {
			arrives: now.Add(-3 * time.Minute),
			want:    "0",
		},
		"zero arrival time": {
			arrives: time.Time{},
			want:    "0",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := minutesAway(tt.arrives, now)
			if got != tt.want {
				t.Errorf("expected %q but got %q", tt.want, got)
			}
		})
	}
}
