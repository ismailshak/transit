package data

type Migration struct {
	Id         int
	Name       string
	MigratedAt string
}

// Base struct holding common fields for database entities
type Entity struct {
	// The table's row id
	Id int
	// When the data was first inserted into the database
	CreatedAt string
	// When the data was last updated in the database
	UpdatedAt string
}

type Location struct {
	Entity
	// The shorthand used to refer to this location. This is the value set in a user's config file
	Slug string
	Name string
}

type Stop struct {
	Entity
	// The ID used when referring to this stop via an API call.
	// Not always the same value as the stop_id in a GTFS
	StopId string
	// The official rider-facing name for the stop
	Name string
	// The location's slug as defined by Location.Slug
	Location string
	// The stop's latitude
	Latitude  string
	Longitude string
}
