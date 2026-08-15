package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/logger"
	_ "modernc.org/sqlite"
)

type TransitDB struct {
	// Exposing the direct database connection if needed
	// but queries and mutations should be made through methods on this struct
	DB *sql.DB

	log *logger.Logger
}

// NewDB opens the SQLite database at path. The file is created if it doesn't
// exist, and no connection is made until the first query.
func NewDB(path string, log *logger.Logger) (*TransitDB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db := &TransitDB{
		DB:  conn,
		log: log,
	}

	return db, nil
}

// Close closes the database connection, once active queries have finished.
func (t *TransitDB) Close() error {
	return t.DB.Close()
}

// Keep migrations up-to-date, and handle first time migration run
func (t *TransitDB) SyncMigrations(ctx context.Context) error {
	err := CreateMigrationTable(ctx, t.DB)
	if err != nil {
		return err
	}

	count, err := GetMigrationCount(ctx, t.DB)
	if err != nil {
		return err
	}

	if count == len(migrationChangesets) {
		return nil
	}

	err = RunMigrations(ctx, t.DB, t.log, count)
	if err != nil {
		return err
	}

	return nil
}

func (t *TransitDB) InsertAgencies(ctx context.Context, agencies []*Agency) error {
	trx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(trx, t.log)

	stmt, err := trx.PrepareContext(ctx, INSERT_AGENCY)
	if err != nil {
		return err
	}

	for _, agency := range agencies {
		_, err = stmt.ExecContext(ctx, agency.AgencyID, agency.Name, agency.Location, agency.Timezone, agency.Language)
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

func (t *TransitDB) GetLocation(ctx context.Context, location LocationSlug) (*Location, error) {
	row := t.DB.QueryRowContext(ctx, SELECT_LOCATION, location)

	var l Location

	err := row.Scan(&l.ID, &l.Slug, &l.Name, &l.SupportsGTFS, &l.CreatedAt, &l.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("scan location: %w", err)
	}

	return &l, nil
}

func (t *TransitDB) GetAllLocations(ctx context.Context) ([]Location, error) {
	rows, err := t.DB.QueryContext(ctx, SELECT_ALL_LOCATIONS)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}

	defer rows.Close()

	locations := make([]Location, 0, 4)

	for rows.Next() {
		var row Location
		if err := rows.Scan(&row.ID, &row.Slug, &row.Name, &row.SupportsGTFS, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		locations = append(locations, row)
	}

	// Catches Next's error during iteration
	return locations, rows.Err()
}

func (t *TransitDB) GetStopsByLocation(ctx context.Context, location LocationSlug, parentsOnly bool) ([]*Stop, error) {
	var statement string
	if parentsOnly {
		statement = SELECT_PARENT_STOPS_BY_LOCATION
	} else {
		statement = SELECT_STOPS_BY_LOCATION
	}

	rows, err := t.DB.QueryContext(ctx, statement, location)
	if err != nil {
		return nil, fmt.Errorf("query stops: %w", err)
	}

	defer rows.Close()

	stops := make([]*Stop, 0, 64) // arbitrary capacity to avoid excessive reallocations

	for rows.Next() {
		var row Stop
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("scan stop: %w", err)
		}

		stops = append(stops, &row)
	}

	return stops, rows.Err()
}

func (t *TransitDB) InsertStops(ctx context.Context, stops []*Stop) error {
	trx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(trx, t.log)

	stmt, err := trx.PrepareContext(ctx, INSERT_STOP)
	if err != nil {
		return err
	}

	for _, stop := range stops {
		_, err = stmt.ExecContext(ctx, stop.StopID, stop.Name, stop.Location, stop.AgencyID, stop.Latitude, stop.Longitude, stop.Type, stop.ParentID)
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

func (t *TransitDB) CountStopsByLocation(ctx context.Context, location LocationSlug) (int, error) {
	row := t.DB.QueryRowContext(ctx, COUNT_STOPS_BY_LOCATION, location)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("scan stop count: %w", err)
	}

	return count, nil
}

func (t *TransitDB) GetLocationAgencies(ctx context.Context, location LocationSlug) ([]Agency, error) {
	rows, err := t.DB.QueryContext(ctx, SELECT_AGENCIES_BY_LOCATION, location)
	if err != nil {
		return nil, fmt.Errorf("query agencies: %w", err)
	}

	defer rows.Close()

	agencies := make([]Agency, 0, 4) // arbitrary capacity to avoid excessive reallocations

	for rows.Next() {
		var row Agency
		if err := rows.Scan(
			&row.ID,
			&row.AgencyID,
			&row.Name,
			&row.Location,
			&row.Timezone,
			&row.Language,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan agency: %w", err)
		}

		agencies = append(agencies, row)
	}

	return agencies, rows.Err()
}

// rollback undoes trx unless it already committed, which reports sql.ErrTxDone.
// Any other error means the rollback itself failed.
func rollback(trx *sql.Tx, log *logger.Logger) {
	err := trx.Rollback()
	if err == nil || errors.Is(err, sql.ErrTxDone) {
		return
	}

	log.Debug(fmt.Sprintf("Failed to roll back transaction: %s", err))
}
