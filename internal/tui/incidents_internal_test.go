package tui

import (
	"testing"
	"time"
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
			date: time.Date(2023, time.June, 10, 7, 27, 0, 0, time.UTC),
			want: "as of 10 Jun 23 7:27am",
		},
		"afternoon, double-digit day": {
			date: time.Date(2029, time.October, 11, 15, 1, 0, 0, time.UTC),
			want: "as of 11 Oct 29 3:01pm",
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
