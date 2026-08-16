package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/logger"
)

func CreateMigrationTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, CreateMigrationsTableSQL)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	return nil
}

func GetMigrationCount(ctx context.Context, db *sql.DB) (int, error) {
	row := db.QueryRowContext(ctx, CountMigrationsSQL)

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

func RunMigrations(ctx context.Context, db *sql.DB, log *logger.Logger, rowCount int) error {
	migrationRows, err := GetCurrentMigrations(ctx, db, rowCount)
	if err != nil {
		return err
	}

	for i, changeset := range migrationChangesets {
		if i+1 > len(migrationRows) {
			err = run(ctx, db, log, &changeset)
			if err != nil {
				return err
			}

			continue
		}

		if changeset.Name != migrationRows[i].Name {
			return errors.New("corrupt migrations (out of sync)")
		}
	}

	return nil
}

func GetCurrentMigrations(ctx context.Context, db *sql.DB, rowCount int) ([]Migration, error) {
	if rowCount == 0 {
		return []Migration{}, nil
	}

	rows, err := db.QueryContext(ctx, SelectMigrationsSQL)
	if err != nil {
		return nil, fmt.Errorf("query migrations: %w", err)
	}

	defer rows.Close()

	migrationRows := make([]Migration, 0, rowCount)

	for rows.Next() {
		var row Migration
		err = rows.Scan(&row.ID, &row.Name, &row.MigratedAt)
		if err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}

		migrationRows = append(migrationRows, row)
	}

	return migrationRows, rows.Err()
}

func run(ctx context.Context, db *sql.DB, log *logger.Logger, changeset *MigrationChangeset) error {
	trx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer rollback(trx, log)

	log.Debug(fmt.Sprintf("Running new database migration: %s", changeset.Name))

	err = changeset.Up(ctx, trx)
	if err != nil {
		return err
	}

	_, err = trx.ExecContext(ctx, InsertMigrationSQL, changeset.Name)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = trx.Commit(); err != nil {
		return err
	}

	return nil
}
