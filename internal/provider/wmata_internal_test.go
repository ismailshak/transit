package provider

import (
	"slices"
	"testing"
	"time"
)

func TestParseLinesAffected(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lines string
		want  []string
	}{
		"single line affected": {
			lines: "RD;",
			want:  []string{"RD"},
		},
		"multiple lines affected": {
			lines: "RD; YL; GR; OR; SV; BL;",
			want:  []string{"RD", "YL", "GR", "OR", "SV", "BL"},
		},
		"no lines": {
			lines: "",
			want:  []string{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := parseLinesAffected(tc.lines)

			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestFormatWMATAStopId(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id   string
		want []string
	}{
		"no id": {
			id:   "",
			want: []string{},
		},
		"id with single platform": {
			id:   "STN_K01",
			want: []string{"K01"},
		},
		"id with multiple platforms": {
			id:   "STN_A01_C01",
			want: []string{"A01", "C01"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := formatWMATAStopID(tc.id)
			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestRailArrival(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 21, 9, 30, 0, 0, time.UTC)

	tests := map[string]struct {
		min          string
		expectedTime time.Time
		expectedOk   bool
	}{
		"boarding now": {
			min:          "BRD",
			expectedTime: asOf,
			expectedOk:   true,
		},
		"arriving": {
			min:          "ARR",
			expectedTime: asOf.Add(30 * time.Second),
			expectedOk:   true,
		},
		"minutes away": {
			min:          "12",
			expectedTime: asOf.Add(12 * time.Minute),
			expectedOk:   true,
		},
		"zero minutes away": {
			min:          "0",
			expectedTime: asOf,
			expectedOk:   true,
		},
		"no prediction": {
			min:          "---",
			expectedTime: time.Time{},
			expectedOk:   false,
		},
		"empty countdown": {
			min:          "",
			expectedTime: time.Time{},
			expectedOk:   false,
		},
		"unrecognized status": {
			min:          "DLY",
			expectedTime: time.Time{},
			expectedOk:   false,
		},
		"negative minutes": {
			min:          "-2",
			expectedTime: asOf.Add(-2 * time.Minute),
			expectedOk:   true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := railArrival(tc.min, asOf)
			if ok != tc.expectedOk {
				t.Errorf("expected ok to be %v but got %v", tc.expectedOk, ok)
			}
			if got != tc.expectedTime {
				t.Errorf("expected time to be %v but got %v", tc.expectedTime, got)
			}
		})
	}
}
