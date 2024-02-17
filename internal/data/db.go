package data

import (
	"database/sql"
	"path/filepath"

	"github.com/ismailshak/transit/internal/config"
	_ "modernc.org/sqlite"
)

// Singleton DB connection throughout execution
var db *TransitDB

type TransitDB struct {
	// Exposing the direct database connection if needed
	// but queries and mutations should be made through methods on this struct
	DB *sql.DB
}

func GetDB() (*TransitDB, error) {
	if db != nil {
		return db, nil
	}

	configPath, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(configPath, "transit.db")
	newDb, err := NewTransitDB(dbPath)

	if err != nil {
		return nil, err
	}

	db = newDb

	return db, nil
}

// Keep migrations up-to-date, and handle first time migration run
func (t *TransitDB) SyncMigrations() error {
	err := CreateMigrationTable(t.DB)
	if err != nil {
		return err
	}

	count, err := GetMigrationCount(t.DB)
	if err != nil {
		return err
	}

	if count == len(migrationChangesets) {
		return nil
	}

	err = RunMigrations(t.DB, count)
	if err != nil {
		return err
	}

	return nil
}

func (t *TransitDB) InsertAgencies(agencies []*Agency) error {
	trx, err := t.DB.Begin()
	if err != nil {
		return err
	}

	// Defer a rollback in case anything fails.
	// Will no-op if Commit succeeds
	defer trx.Rollback()

	stmt, err := trx.Prepare(INSERT_AGENCY)
	if err != nil {
		return err
	}

	for _, agency := range agencies {
		_, err = stmt.Exec(agency.AgencyID, agency.Name, agency.Location, agency.Timezone, agency.Language)
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

func (t *TransitDB) GetLocation(location LocationSlug) (*Location, error) {
	row := t.DB.QueryRow(SELECT_LOCATION, location)

	var l Location

	err := row.Scan(&l.ID, &l.Slug, &l.Name, &l.SupportsGTFS, &l.CreatedAt, &l.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &l, nil
}

func (t *TransitDB) GetStopsByLocation(location LocationSlug, parentsOnly bool) ([]*Stop, error) {
	var statement string
	if parentsOnly {
		statement = SELECT_PARENT_STOPS_BY_LOCATION
	} else {
		statement = SELECT_STOPS_BY_LOCATION
	}

	rows, err := t.DB.Query(statement, location)
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
			&row.AgencyID,
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

func (t *TransitDB) InsertStops(stops []*Stop) error {
	trx, err := t.DB.Begin()
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
		_, err = stmt.Exec(stop.StopID, stop.Name, stop.Location, stop.AgencyID, stop.Latitude, stop.Longitude, stop.Type, stop.ParentID)
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

func (t *TransitDB) CountStopsByLocation(location LocationSlug) (int, error) {
	row := t.DB.QueryRow(COUNT_STOPS_BY_LOCATION, location)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Exists for testing purposes. Use GetDB instead
func NewTransitDB(path string) (*TransitDB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db := &TransitDB{
		DB: conn,
	}

	return db, nil
}
