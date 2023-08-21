package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ismailshak/transit/logger"
)

func createMigrationTable(db *sql.DB) {
	stmt, err := db.Prepare(CREATE_MIGRATIONS_TABLE)

	// TODO: better error handling
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = stmt.Exec()

	if err != nil {
		fmt.Println(err)
		return
	}

}

func latestMigrations(db *sql.DB) {
	hasLatest, count := hasLatestMigration(db)
	if hasLatest {
		return
	}

	logger.Info("New database updates found. Syncing database before executing command")

	runMigrations(db, count)
}

func hasLatestMigration(db *sql.DB) (bool, int) {
	row := db.QueryRow(MIGRATIONS_COUNT)

	var count int
	row.Scan(&count)

	return count == len(migrationChangesets), count
}

func runMigrations(db *sql.DB, rowCount int) {
	migrationRows := getCurrentMigrations(db, rowCount)

	for i, changeset := range migrationChangesets {
		if i+1 > len(migrationRows) {
			run(db, &changeset)
			continue
		}

		if changeset.Name != migrationRows[i].Name {
			// TODO: handle error
			panic("Corrupt migrations")
		}
	}
}

func getCurrentMigrations(db *sql.DB, rowCount int) []Migration {
	if rowCount == 0 {
		return []Migration{}
	}

	rows, err := db.Query(SELECT_MIGRATIONS)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer rows.Close()

	migrationRows := make([]Migration, 0, rowCount)

	for rows.Next() {
		var row Migration
		rows.Scan(&row.Id, &row.Name, &row.MigratedAt)
		migrationRows = append(migrationRows, row)
	}

	return migrationRows
}

func run(db *sql.DB, changeset *MigrationChangeset) {
	trx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Defer a rollback in case anything fails.
	// Will no-op of Commit succeeds
	defer trx.Rollback()

	logger.Debug(fmt.Sprintf("Running new database migration: %s", changeset.Name))

	ok := changeset.Up(context.Background(), trx)

	if !ok {
		return
	}

	ok = insertMigrationRecord(context.Background(), trx, changeset.Name)

	if !ok {
		// TODO: exit process here
		return
	}

	// Commit the transaction.
	if err = trx.Commit(); err != nil {
		fmt.Println(err)
		// TODO: exit process here?
		return
	}
}

func insertMigrationRecord(ctx context.Context, trx *sql.Tx, name string) bool {
	_, err := trx.ExecContext(ctx, INSERT_MIGRATION, name)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to insert handler: %s", name))
		fmt.Println(err)
		return false
	}

	return true
}
