package data

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func getTestDb(t *testing.T) *sql.DB {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "runner_test.db")

	t.Logf("Temp database at: %s", dbPath)

	db, err := DbConnect(dbPath)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func initMigrationsTable(t *testing.T, db *sql.DB) {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %s", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(CREATE_MIGRATIONS_TABLE)
	if err != nil {
		t.Fatalf("Failed to create migration: %s", err)
	}

	_, err = tx.Exec(INSERT_MIGRATION, "1_FakeMigration")

	tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %s", err)
	}
}

func TestCreatingMigrationTable(t *testing.T) {
	t.Parallel()

	db := getTestDb(t)

	err := createMigrationTable(db)
	if err != nil {
		t.Errorf("Failed to create migrations table: %s", err)
	}
}

func TestSkippingMigrationsTableIfExists(t *testing.T) {
	t.Parallel()

	db := getTestDb(t)
	initMigrationsTable(t, db)

	err := createMigrationTable(db)
	if err != nil {
		t.Errorf("Failed to skipping migrations table: %s", err)
	}

	count, err := getMigrationCount(db)
	if err != nil {
		t.Errorf("Failed to get migration count: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 migration. Got %q", count)
	}

	migrations, err := getCurrentMigrations(db, count)
	if err != nil {
		t.Errorf("Failed to get migrations: %s", err)
	}

	if migrations[0].Id != 1 {
		t.Errorf("Expected migration to have ID \"1\". Got %q", migrations[0].Id)
	}

	if migrations[0].Name != "1_FakeMigration" {
		t.Errorf("Expected migration \"1_FakeMigration\". Got %q", migrations[0].Name)
	}

	if migrations[0].MigratedAt == "" {
		t.Errorf("Expected migration to have a valid date. Got %q", migrations[0].MigratedAt)
	}
}
