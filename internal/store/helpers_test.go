package store_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/store"
)

// blankDB creates a temporary database that has no tables
func blankDB(t *testing.T) *store.Store {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-blank.db")

	t.Logf("Temp database at: %s", dbPath)

	db, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("expected no error but got %v, the temp database is still open", err)
		}
	})

	return db
}

// migratedDB creates a temporary test database that's fully migrated
func migratedDB(t *testing.T) *store.Store {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-migrated.db")

	db, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	if err = db.SyncMigrations(t.Context()); err != nil {
		t.Fatal("Failed to migrate test database", err)
	}

	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("expected no error but got %v, the temp database is still open", err)
		}
	})

	return db
}

func initMigrationsTable(t *testing.T, db *sql.DB) {
	t.Helper()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %s", err)
	}

	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(t.Context(), store.CreateMigrationsTableSQL)
	if err != nil {
		t.Fatalf("Failed to create migration: %s", err)
	}

	_, err = tx.ExecContext(t.Context(), store.InsertMigrationSQL, "1_FakeMigration")
	if err != nil {
		t.Fatalf("Failed to insert migration: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %s", err)
	}
}
