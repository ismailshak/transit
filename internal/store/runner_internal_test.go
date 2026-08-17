package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// openTestDB opens a temporary database that has no tables
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "transit-test-runner.db")

	s, err := NewStore(path)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	t.Cleanup(func() {
		err := s.Close()
		if err != nil {
			t.Errorf("expected no error but got %v, the temp database is still open", err)
		}
	})

	return s.db
}

func initMigrationsTable(t *testing.T, db *sql.DB) {
	t.Helper()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %s", err)
	}

	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(t.Context(), createMigrationsTableSQL)
	if err != nil {
		t.Fatalf("Failed to create migration: %s", err)
	}

	_, err = tx.ExecContext(t.Context(), insertMigrationSQL, "1_FakeMigration")
	if err != nil {
		t.Fatalf("Failed to insert migration: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %s", err)
	}
}

func TestCreatingMigrationTable(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)

	err := createMigrationTable(t.Context(), db)
	if err != nil {
		t.Errorf("Failed to create migrations table: %s", err)
	}
}

func TestSkippingMigrationsTableIfExists(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	initMigrationsTable(t, db)

	err := createMigrationTable(t.Context(), db)
	if err != nil {
		t.Errorf("Failed skipping migrations table: %s", err)
	}

	count, err := migrationCount(t.Context(), db)
	if err != nil {
		t.Errorf("Failed to get migration count: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 migration. Got %d", count)
	}

	migrations, err := currentMigrations(t.Context(), db, count)
	if err != nil {
		t.Errorf("Failed to get migrations: %s", err)
	}

	if migrations[0].ID != 1 {
		t.Errorf("Expected migration to have ID 1. Got %d", migrations[0].ID)
	}

	if migrations[0].Name != "1_FakeMigration" {
		t.Errorf("Expected migration name \"1_FakeMigration\". Got %q", migrations[0].Name)
	}

	if migrations[0].MigratedAt == "" {
		t.Errorf("Expected migration to have a valid date. Got %q", migrations[0].MigratedAt)
	}
}
