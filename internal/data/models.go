package data

// The unique identifier for a location
type LocationSlug string

const (
	DMVSlug LocationSlug = "dmv"
	SFSlug  LocationSlug = "sf"
)

// Used in the database to differentiate between the different types
type StopType string

const (
	TrainStation StopType = "train" // Type used to represent a train station
	BusStop      StopType = "bus"   // Type used to represent a bus stop
)

// A record of a database migration that was executed
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

// A public entity administrating and managing transit services
type Agency struct {
	Entity
	// Identifies a transit brand which is often synonymous with a transit agency
	AgencyID string
	// Full name of the transit agency
	Name string
	// The location's slug as defined by Location.Slug
	Location LocationSlug
	// Timezone where the transit agency is located.
	// Usually a TZ timezone from the https://www.iana.org/time-zones
	Timezone string
	// Primary language used by this transit agency.
	// Usually a code from https://www.w3.org/International/articles/language-tags/
	Language string
}

// A geographical location in the world where a transit agency is operating
type Location struct {
	Entity
	// The shorthand used to refer to this location. This is the value set in a user's config file
	Slug LocationSlug
	// Rider-facing name
	Name string
	// Whether the API behind it supports GTFS data
	SupportsGTFS bool
}

// A place where vehicles pick up or drop off riders
type Stop struct {
	Entity
	// The official ID of this stop
	StopID string
	// The official rider-facing name for the stop
	Name string
	// A FK to the Location's `Slug`
	Location LocationSlug
	// A FK to the agency's `AgencyID`
	AgencyID string
	// The stop's latitude
	Latitude string
	// The stop's longitude
	Longitude string
	// "train" | "bus"
	Type StopType
	// A StopID if this stop is embedded inside another
	ParentID string
}

// Wrapper type that implements the fuzzy matching interface to enable search
type SearchableStops []*Stop

func (s SearchableStops) Len() int            { return len(s) }
func (s SearchableStops) String(i int) string { return s[i].Name }
