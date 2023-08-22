package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ismailshak/transit/logger"
)

func createMigrationTable(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), CREATE_MIGRATIONS_TABLE)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %s", err)
	}

	return nil
}

func getMigrationCount(db *sql.DB) (int, error) {
	row := db.QueryRow(MIGRATIONS_COUNT)

	var count int
	err := row.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

func runMigrations(db *sql.DB, rowCount int) error {
	migrationRows, err := getCurrentMigrations(db, rowCount)
	if err != nil {
		return err
	}

	for i, changeset := range migrationChangesets {
		if i+1 > len(migrationRows) {
			err = run(db, &changeset)
			if err != nil {
				return err
			}

			continue
		}

		if changeset.Name != migrationRows[i].Name {
			return fmt.Errorf("corrupt migrations (out of sync)")
		}
	}

	return nil
}

func getCurrentMigrations(db *sql.DB, rowCount int) ([]Migration, error) {
	if rowCount == 0 {
		return []Migration{}, nil
	}

	rows, err := db.Query(SELECT_MIGRATIONS)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch database migrations. %s", err)
	}

	defer rows.Close()

	migrationRows := make([]Migration, 0, rowCount)

	for rows.Next() {
		var row Migration
		err = rows.Scan(&row.ID, &row.Name, &row.MigratedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration row. %s", err)
		}

		migrationRows = append(migrationRows, row)
	}

	return migrationRows, nil
}

func run(db *sql.DB, changeset *MigrationChangeset) error {
	trx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin database transaction. %s", err)
	}

	// Defer a rollback in case anything fails.
	// Will no-op if Commit succeeds
	defer trx.Rollback()

	logger.Debug(fmt.Sprintf("Running new database migration: %s", changeset.Name))

	err = changeset.Up(context.Background(), trx)

	if err != nil {
		return err
	}

	_, err = trx.ExecContext(context.Background(), INSERT_MIGRATION, changeset.Name)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = trx.Commit(); err != nil {
		return err
	}

	return nil
}
