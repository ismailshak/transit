package data

// The unique identifier for a location
type LocationSlug string

const (
	DMVSlug LocationSlug = "dmv"
)

// Used in the database to differentiate between the different types
type StopType string

const (
	TrainStation StopType = "train" // Type used to represent a train station
	BusStop      StopType = "bus"   // Type used to represent a bus stop
)

type Migration struct {
	ID         int
	Name       string
	MigratedAt string
}

// Base struct holding common fields for database entities
type Entity struct {
	// The table's row id
	ID int
	// When the data was first inserted into the database
	CreatedAt string
	// When the data was last updated in the database
	UpdatedAt string
}

type Location struct {
	Entity
	// The shorthand used to refer to this location. This is the value set in a user's config file
	Slug LocationSlug
	// Rider-facing name
	Name string
	// Whether the API behind it supports GTFS data
	SupportsGTFS bool
}

type Stop struct {
	Entity
	// The official ID of this stop
	StopID string
	// The official rider-facing name for the stop
	Name string
	// The location's slug as defined by Location.Slug
	Location LocationSlug
	// The stop's latitude
	Latitude string
	// The stop's longitude
	Longitude string
	// "train" | "bus"
	Type StopType
	// A StopID if this stop is embedded inside another
	ParentID string
}

// Wrapper type to satisfy the fuzzy matching interface
type SearchableStops []*Stop

func (s SearchableStops) Len() int            { return len(s) }
func (s SearchableStops) String(i int) string { return s[i].Name }
