package transit

import "time"

// LocationSlug is the unique identifier for a location.
type LocationSlug string

const (
	DMVSlug LocationSlug = "dmv"
	SFSlug  LocationSlug = "sf"
)

// StopType is used in the database to differentiate between the different types.
type StopType string

// Mode is the type of vehicle that serves a stop.
type Mode string

const (
	ModeBus   Mode = "bus"
	ModeFerry Mode = "ferry"
	ModeMetro Mode = "metro"
	ModeRail  Mode = "rail"
)

// RefKind is the type of entity an [AlertRef] affects.
type RefKind string

const (
	RefAgency RefKind = "agency"
	RefRoute  RefKind = "route"
	RefStop   RefKind = "stop"
)

const (
	TrainStation StopType = "train" // Type used to represent a train station.
	BusStop      StopType = "bus"   // Type used to represent a bus stop.
)

// StoreEntity is the base struct holding common fields for database entities.
type StoreEntity struct {
	// The table's row id.
	ID int
	// When the data was first inserted into the database.
	CreatedAt string
	// When the data was last updated in the database.
	UpdatedAt string
}

// StopRef identifies one stop by the ID a Source answers on. The resolve path derives it from a
// [Stop]. One Stop can produce several StopRefs when a Source addresses parts of a stop separately.
type StopRef struct {
	StopID   string // ID the Source answers on. Not always the one the stop was seeded with.
	Name     string // Rider-facing name.
	AgencyID string // The agency that operates service at this stop. Some sources need it to build the request.
	Source   string // Provider source that answers for this stop. A stop belongs to exactly one.
}

// Departure is one upcoming vehicle at a stop.
type Departure struct {
	Source    string // Provider source that gave us the Departure data. It joins the row to its SourceStatus.
	StopID    string // ID used by the Source to identify the stop.
	StopName  string // Rider-facing name for the stop.
	AgencyID  string // The agency that operates this vehicle.
	Mode      Mode
	Line      string    // Rider-facing name for the line.
	LineColor string    // The Line's background color.
	LineText  string    // The Line's foreground color.
	Headsign  string    // Rider-facing destination displayed.
	Direction string    // Which way along the line the vehicle travels. Departures at a stop are grouped by it.
	Arrives   time.Time // Always an absolute instant. Render it in the agency's zone.
}

// SourceStatus is the outcome of asking one source for data.
type SourceStatus struct {
	Source string    // Provider source that was asked.
	AsOf   time.Time // When the source produced the data, not when we asked for it.
	Err    error     // Non-nil means this particular Source is degraded. The others may still be good.
}

// DepartureSet is the result of asking every source that serves the requested stops. The status
// of each source travels with the rows, so a combined result can report its age and its failures.
type DepartureSet struct {
	Departures []Departure    // Departures that came back for the requested stops.
	Sources    []SourceStatus // One entry per source that was asked.
}

// AsOf returns the time of the oldest source that succeeded.
// It returns the zero time when no source succeeded.
func (s DepartureSet) AsOf() time.Time { return oldest(s.Sources) }

// Degraded returns the sources that failed. It's empty when all sources succeed.
func (s DepartureSet) Degraded() []SourceStatus { return degraded(s.Sources) }

// AlertRef is the entity an Alert applies to.
type AlertRef struct {
	Kind      RefKind
	ID        string // ID used by the Source to identify the entity.
	Color     string // The entity's background color. Empty for entities without branding (e.g. stops).
	TextColor string // The entity's foreground color. Empty for entities without branding (e.g. stops).
}

// Alert is a disruption to service (planned or not) and the entities it applies to.
type Alert struct {
	Source      string // Provider source that produced the data.
	AgencyID    string // The agency whose service is disrupted.
	Affected    []AlertRef
	Description string    // Rider-facing text describing the disruption.
	Effect      string    // The agency's label for the effect this alert has on riders.
	Starts      time.Time // Zero means the alert is already in effect.
	Ends        time.Time // Zero means no announced end.
	Updated     time.Time // When the Source last updated the alert. Zero if it didn't publish one.
}

// AlertSet is the result of asking every source for the disruptions it knows about.
type AlertSet struct {
	Alerts  []Alert
	Sources []SourceStatus // One entry per source that was asked.
}

// AsOf returns the time of the oldest source that succeeded, which is the honest age of the whole
// set. It returns the zero time when no source succeeded.
func (s AlertSet) AsOf() time.Time { return oldest(s.Sources) }

// Degraded returns the sources that failed. It's empty for the happy path.
func (s AlertSet) Degraded() []SourceStatus { return degraded(s.Sources) }

// Route is a line that vehicles run along. It carries the line's display identity, which a
// departure resolves through the reference the source gives for it.
type Route struct {
	RouteID   string // ID used by the Source to identify this route.
	ShortName string // User-facing short name for the line.
	Color     string // The line's background color.
	TextColor string // The line's foreground color.
	Mode      Mode
	AgencyID  string // The agency that operates this route.
}

// Trip is a single journey of one vehicle along a route, and the destination it displays while
// making that journey.
type Trip struct {
	TripID   string // ID used by the Source to identify this trip.
	RouteID  string // The route this trip runs along.
	Headsign string // User-facing destination/direction displayed.
	ShapeID  string // The physical path the vehicle follows. Trips that run the same way share one.
}

// StopRoute records that a route serves a stop.
type StopRoute struct {
	StopID  string
	RouteID string
}

// Agency is a public entity administrating and managing transit services.
type Agency struct {
	StoreEntity
	AgencyID string       // Identifies a transit brand which is often synonymous with a transit agency.
	Name     string       // Full name of the transit agency.
	Location LocationSlug // The location's slug as defined by Location.Slug.
	// Timezone where the transit agency is located.
	// Usually a TZ timezone from https://www.iana.org/time-zones.
	Timezone string
	// Primary language used by this transit agency.
	// Usually a code from https://www.w3.org/International/articles/language-tags/.
	Language string
}

// Location is a geographical location in the world where a transit agency is operating.
type Location struct {
	StoreEntity
	Slug         LocationSlug // The shorthand used to refer to this location. This is the value set in a user's config file.
	Name         string       // Rider-facing name.
	SupportsGTFS bool         // Whether the API behind it supports GTFS data.
}

// Stop is a place where vehicles pick up or drop off riders. Its StopID is the one the seed
// published. [StopRef] carries the ID a Source answers on (which can be different).
type Stop struct {
	StoreEntity
	StopID    string       // The seeded ID of this stop.
	Name      string       // The seeded rider-facing name for the stop.
	Location  LocationSlug // A FK to the Location's `Slug`.
	AgencyID  string       // A FK to the agency's `AgencyID`.
	Latitude  string       // The stop's latitude.
	Longitude string       // The stop's longitude.
	Type      StopType
	ParentID  string // A StopID if this stop is embedded inside another.
}

// Static is the reference data a Source seeds. This is everything that doesn't change between
// fetches. A Source omits what it has no equivalent for.
type Static struct {
	Agencies   []Agency
	Stops      []Stop
	Routes     []Route
	Trips      []Trip
	StopRoutes []StopRoute
	Version    string // Identifies this edition of the data.
}

func oldest(sources []SourceStatus) time.Time {
	var t time.Time
	for _, src := range sources {
		if src.Err == nil && (t.IsZero() || src.AsOf.Before(t)) {
			t = src.AsOf
		}
	}

	return t
}

func degraded(sources []SourceStatus) []SourceStatus {
	var bad []SourceStatus
	for _, src := range sources {
		if src.Err != nil {
			bad = append(bad, src)
		}
	}

	return bad
}
