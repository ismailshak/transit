package transit_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/transit"
)

var errSourceDown = errors.New("401 Unauthorized")

// What a source that fans out reports when it loses some of its requests but not all of them.
var errSomeStopsLost = errors.Join(
	errors.New(`departures at "Embarcadero": 503 Service Unavailable`),
	errors.New(`departures at "Powell St": 503 Service Unavailable`),
)

var (
	nine       = time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	nineOhFive = time.Date(2026, 8, 20, 9, 5, 0, 0, time.UTC)
	nineTen    = time.Date(2026, 8, 20, 9, 10, 0, 0, time.UTC)
)

func TestAsOf(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sources  []transit.SourceStatus
		expected time.Time
	}{
		"no sources were asked": {
			sources:  nil,
			expected: time.Time{},
		},
		"one source succeeded": {
			sources:  []transit.SourceStatus{{Source: "wmata-rail", AsOf: nine}},
			expected: nine,
		},
		"the oldest of the sources that returned data": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", AsOf: nineOhFive},
				{Source: "wmata-bus", AsOf: nine},
				{Source: "art", AsOf: nineTen},
			},
			expected: nine,
		},
		"a source that lost some of its requests is included": {
			sources: []transit.SourceStatus{
				{Source: "bayarea511", AsOf: nine, Err: errSomeStopsLost},
				{Source: "wmata-rail", AsOf: nineOhFive},
			},
			expected: nine,
		},
		"a source that returned nothing is excluded": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", Err: errSourceDown},
				{Source: "art", AsOf: nineOhFive},
			},
			expected: nineOhFive,
		},
		"all sources failed": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", Err: errSourceDown},
				{Source: "art", Err: errSourceDown},
			},
			expected: time.Time{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			departures := transit.DepartureSet{Sources: tc.sources}
			if got := departures.AsOf(); !got.Equal(tc.expected) {
				t.Errorf("expected %v but got %v", tc.expected, got)
			}

			alerts := transit.AlertSet{Sources: tc.sources}
			if got := alerts.AsOf(); !got.Equal(tc.expected) {
				t.Errorf("expected %v but got %v", tc.expected, got)
			}
		})
	}
}

func TestDegraded(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sources  []transit.SourceStatus
		expected []string
	}{
		"no sources were asked": {
			sources:  nil,
			expected: nil,
		},
		"every source succeeded": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", AsOf: nine},
				{Source: "art", AsOf: nineOhFive},
			},
			expected: nil,
		},
		"one source of three failed": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", AsOf: nine},
				{Source: "wmata-bus", Err: errSourceDown},
				{Source: "art", AsOf: nineTen},
			},
			expected: []string{"wmata-bus"},
		},
		"a source that returned data is still degraded when it lost some of it": {
			sources: []transit.SourceStatus{
				{Source: "bayarea511", AsOf: nine, Err: errSomeStopsLost},
			},
			expected: []string{"bayarea511"},
		},
		"all sources failed": {
			sources: []transit.SourceStatus{
				{Source: "wmata-rail", Err: errSourceDown},
				{Source: "art", Err: errSourceDown},
			},
			expected: []string{"wmata-rail", "art"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			departures := transit.DepartureSet{Sources: tc.sources}
			assertDegraded(t, departures.Degraded(), tc.expected)

			alerts := transit.AlertSet{Sources: tc.sources}
			assertDegraded(t, alerts.Degraded(), tc.expected)
		})
	}
}

func assertDegraded(t *testing.T, got []transit.SourceStatus, expected []string) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("expected %d degraded sources but got %d", len(expected), len(got))
	}

	for i, want := range expected {
		if got[i].Source != want {
			t.Errorf("expected %q but got %q", want, got[i].Source)
		}

		if got[i].Err == nil {
			t.Errorf("expected an error on %q but got none", got[i].Source)
		}
	}
}
