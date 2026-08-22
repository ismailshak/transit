package transit

import (
	"context"
	"errors"
)

// ErrNoDepartures is returned when a stop has no upcoming departures.
var ErrNoDepartures = errors.New("no departures")

// Provider is the set of calls a command makes to a transit source.
type Provider interface {
	// Departures returns what's arriving at the given refs.
	Departures(ctx context.Context, refs []StopRef) (DepartureSet, error)

	// Alerts returns the disruptions the source is reporting.
	Alerts(ctx context.Context) (AlertSet, error)

	// StopRefs turns a seeded stop into the refs this source understands. One stop
	// can be several refs. The IDs aren't always the ones the stop was seeded with.
	StopRefs(stop Stop) []StopRef
}

// Seeder is implemented by sources that can fetch and return their whole stop list.
type Seeder interface {
	// Seed returns the source's whole stop list so that init scan write it to the store.
	Seed(context.Context) (*Static, error)
}
