package data

import (
	"database/sql"
)

// The unique identifier for a location
type LocationSlug string

const (
	DmvSlug LocationSlug = "dmv"
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
	Name string
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

func GetStopsByLocation(db *sql.DB, location LocationSlug, parentsOnly bool) ([]*Stop, error) {
	var statement string
	if parentsOnly {
		statement = SELECT_PARENT_STOPS_BY_LOCATION
	} else {
		statement = SELECT_STOPS_BY_LOCATION
	}

	rows, err := db.Query(statement, location)
	if err != nil {
		return nil, err
	}

	stops := make([]*Stop, 0, 64) // arbitrary capacity to avoid excessive reallocations

	for rows.Next() {
		var row Stop
		rows.Scan(
			&row.ID,
			&row.StopID,
			&row.Name,
			&row.Location,
			&row.Latitude,
			&row.Longitude,
			&row.Type,
			&row.ParentID,
			&row.CreatedAt,
			&row.UpdatedAt,
		)

		stops = append(stops, &row)
	}

	return stops, nil
}

func InsertStops(db *sql.DB, stops []*Stop) error {
	trx, err := db.Begin()
	if err != nil {
		return err
	}

	// Defer a rollback in case anything fails.
	// Will no-op if Commit succeeds
	defer trx.Rollback()

	stmt, err := trx.Prepare(INSERT_STOP)
	if err != nil {
		return err
	}

	for _, stop := range stops {
		_, err = stmt.Exec(stop.StopID, stop.Name, stop.Location, stop.Latitude, stop.Longitude, stop.Type, stop.ParentID)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	if err = trx.Commit(); err != nil {
		return err
	}

	return nil
}

func CountStopsByLocation(db *sql.DB, location LocationSlug) (int, error) {
	row := db.QueryRow(COUNT_STOPS_BY_LOCATION, location)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
