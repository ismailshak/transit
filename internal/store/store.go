// Package store owns the SQLite database on a user's machine. Data
// enters from a provider or from a parsed GTFS feed.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/transit"
	_ "modernc.org/sqlite"
)

type Store struct {
	// Exposing the direct database connection if needed
	// but queries and mutations should be made through methods on this struct
	// TODO: Make private
	DB *sql.DB
}

// NewStore opens the SQLite database at path. The file is created if it doesn't
// exist and no connection is made until the first query.
func NewStore(path string) (*Store, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	return &Store{DB: conn}, nil
}

// Ping verifies the connection to the database.
func (s *Store) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

// Close closes the database connection, once active queries have finished.
func (s *Store) Close() error {
	return s.DB.Close()
}

// SyncMigrations keeps migrations up-to-date and handles first time migration run
func (s *Store) SyncMigrations(ctx context.Context) error {
	err := CreateMigrationTable(ctx, s.DB)
	if err != nil {
		return err
	}

	count, err := MigrationCount(ctx, s.DB)
	if err != nil {
		return err
	}

	if count == len(migrationChangesets) {
		return nil
	}

	err = runMigrations(ctx, s.DB, count)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) InsertAgencies(ctx context.Context, agencies []*transit.Agency) error {
	trx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(trx)

	stmt, err := trx.PrepareContext(ctx, InsertAgencySQL)
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

// Location returns one location by its slug. A slug with no row returns nil.
func (s *Store) Location(ctx context.Context, location transit.LocationSlug) (*transit.Location, error) {
	row := s.DB.QueryRowContext(ctx, SelectLocationSQL, location)

	var l transit.Location

	err := row.Scan(&l.ID, &l.Slug, &l.Name, &l.SupportsGTFS, &l.CreatedAt, &l.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("scan location: %w", err)
	}

	return &l, nil
}

// AllLocations returns every location the migrations seeded.
func (s *Store) AllLocations(ctx context.Context) ([]transit.Location, error) {
	rows, err := s.DB.QueryContext(ctx, SelectAllLocationsSQL)
	if err != nil {
		return nil, fmt.Errorf("query locations: %w", err)
	}

	defer rows.Close()

	locations := make([]transit.Location, 0, 4)

	for rows.Next() {
		var row transit.Location
		if err := rows.Scan(&row.ID, &row.Slug, &row.Name, &row.SupportsGTFS, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}

		locations = append(locations, row)
	}

	// Catches Next's error during iteration
	return locations, rows.Err()
}

// StopsByLocation returns the stops seeded for a location. parentsOnly only returns
// stops with no parent. Those are the stations rather and not the platforms
// underneath them.
func (s *Store) StopsByLocation(ctx context.Context, location transit.LocationSlug, parentsOnly bool) ([]*transit.Stop, error) {
	var statement string
	if parentsOnly {
		statement = SelectParentStopsByLocationSQL
	} else {
		statement = SelectStopsByLocationSQL
	}

	rows, err := s.DB.QueryContext(ctx, statement, location)
	if err != nil {
		return nil, fmt.Errorf("query stops: %w", err)
	}

	defer rows.Close()

	stops := make([]*transit.Stop, 0, 64) // arbitrary capacity to avoid excessive reallocations

	for rows.Next() {
		var row transit.Stop
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

func (s *Store) InsertStops(ctx context.Context, stops []*transit.Stop) error {
	trx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer rollback(trx)

	stmt, err := trx.PrepareContext(ctx, InsertStopSQL)
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

// CountStopsByLocation returns the number of stops seeded for a location slug.
func (s *Store) CountStopsByLocation(ctx context.Context, location transit.LocationSlug) (int, error) {
	row := s.DB.QueryRowContext(ctx, CountStopsByLocationSQL, location)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("scan stop count: %w", err)
	}

	return count, nil
}

// LocationAgencies reads the agencies seeded for a location. An unseeded location
// returns an empty slice and no error. Nothing tells that apart from a location
// with no agencies.
func (s *Store) LocationAgencies(ctx context.Context, location transit.LocationSlug) ([]transit.Agency, error) {
	rows, err := s.DB.QueryContext(ctx, SelectAgenciesByLocationSQL, location)
	if err != nil {
		return nil, fmt.Errorf("query agencies: %w", err)
	}

	defer rows.Close()

	agencies := make([]transit.Agency, 0, 4) // arbitrary capacity to avoid excessive reallocations

	for rows.Next() {
		var row transit.Agency
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

// MatchStops fuzzy matches query against the names of the stations seeded for a
// location. Matches come back best first. No match is an empty slice.
func (s *Store) MatchStops(ctx context.Context, location transit.LocationSlug, query string) ([]*transit.Stop, error) {
	stops, err := s.StopsByLocation(ctx, location, true)
	if err != nil {
		return nil, err
	}

	matches := FuzzyFindFrom(query, SearchableStops(stops))

	matched := make([]*transit.Stop, 0, matches.Len())
	for _, m := range matches {
		matched = append(matched, stops[m.Index])
	}

	return matched, nil
}

// rollback undoes trx unless it already committed which reports sql.ErrTxDone.
// TODO: once there's a debug stream to write to report this rollback error
func rollback(trx *sql.Tx) {
	_ = trx.Rollback()
}
