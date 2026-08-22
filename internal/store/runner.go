package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// migration is a record of a database migration that was executed.
type migration struct {
	ID         int
	Name       string
	MigratedAt string
}

func createMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createMigrationsTableSQL)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	return nil
}

// migrationCount reports how many migrations have already been applied.
func migrationCount(ctx context.Context, db *sql.DB) (int, error) {
	row := db.QueryRowContext(ctx, countMigrationsSQL)

	var count int
	err := row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("scan migration count: %w", err)
	}

	return count, nil
}

func runMigrations(ctx context.Context, db *sql.DB, rowCount int) error {
	migrationRows, err := currentMigrations(ctx, db, rowCount)
	if err != nil {
		return err
	}

	for i, cs := range migrationChangesets {
		if i+1 > len(migrationRows) {
			err = run(ctx, db, &cs)
			if err != nil {
				return err
			}

			continue
		}

		if cs.Name != migrationRows[i].Name {
			return errors.New("corrupt migrations (out of sync)")
		}
	}

	return nil
}

// currentMigrations reads the migrations already applied.
func currentMigrations(ctx context.Context, db *sql.DB, rowCount int) ([]migration, error) {
	if rowCount == 0 {
		return []migration{}, nil
	}

	rows, err := db.QueryContext(ctx, selectMigrationsSQL)
	if err != nil {
		return nil, fmt.Errorf("query migrations: %w", err)
	}

	defer rows.Close()

	migrationRows := make([]migration, 0, rowCount)

	for rows.Next() {
		var row migration
		err = rows.Scan(&row.ID, &row.Name, &row.MigratedAt)
		if err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}

		migrationRows = append(migrationRows, row)
	}

	return migrationRows, rows.Err()
}

func run(ctx context.Context, db *sql.DB, cs *changeset) error {
	trx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer rollback(trx)

	err = cs.Up(ctx, trx)
	if err != nil {
		return err
	}

	_, err = trx.ExecContext(ctx, insertMigrationSQL, cs.Name)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = trx.Commit(); err != nil {
		return err
	}

	return nil
}
