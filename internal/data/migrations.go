package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// A function that will apply a change to the database schema
type Changeset func(ctx context.Context, trx *sql.Tx) error

type MigrationChangeset struct {
	// Name of the migration. Value will be stored in the migrations table
	Name string
	// Applies forward changes to the database schema
	Up Changeset
	// Rolls back the changes made by Up
	Down Changeset
}

// The list of all migrations to run
var migrationChangesets = []MigrationChangeset{
	{
		Name: "0001_Init",
		Up:   createInitialTables,
		Down: dropInitialTables,
	},
	{
		Name: "0002_Add_DMV",
		Up:   addDMVToLocations,
		Down: deleteDMVFromLocations,
	},
	{
		Name: "0003_Add_SF",
		Up:   addSFToLocations,
		Down: deleteSFFromLocations,
	},
}

func failedMigration(message string, err error) error {
	return errors.New(fmt.Sprint(message, err))
}

func createInitialTables(ctx context.Context, trx *sql.Tx) error {
	_, err := trx.ExecContext(ctx, CREATE_LOCATIONS_TABLE)
	if err != nil {
		return failedMigration("failed to create 'locations' table: ", err)
	}

	_, err = trx.ExecContext(ctx, CREATE_AGENCIES_TABLE)
	if err != nil {
		return failedMigration("failed to create 'agencies' table: ", err)
	}

	_, err = trx.ExecContext(ctx, CREATE_STOPS_TABLE)
	if err != nil {
		return failedMigration("failed to create 'stops' table: ", err)
	}

	_, err = trx.ExecContext(ctx, CREATE_STOP_LOCATION_INDEX)
	if err != nil {
		return failedMigration("failed to create 'stop.location' index: ", err)
	}

	return nil
}

func dropInitialTables(ctx context.Context, trx *sql.Tx) error {
	return nil
}

func addDMVToLocations(ctx context.Context, trx *sql.Tx) error {
	_, err := trx.ExecContext(
		ctx,
		INSERT_LOCATION,
		DMVSlug,
		"District Of Columbia, Maryland and Virginia (US)",
		true,
	)

	if err != nil {
		return failedMigration("failed to insert 'dmv' into 'locations': ", err)
	}

	return nil
}

func deleteDMVFromLocations(ctx context.Context, trx *sql.Tx) error {
	return nil
}

func addSFToLocations(ctx context.Context, trx *sql.Tx) error {
	_, err := trx.ExecContext(
		ctx,
		INSERT_LOCATION,
		SFSlug,
		"San Francisco Bay Area (US)",
		true,
	)

	if err != nil {
		return failedMigration("failed to insert 'sf' into 'locations': ", err)
	}

	return nil
}

func deleteSFFromLocations(ctx context.Context, trx *sql.Tx) error {
	return nil
}
