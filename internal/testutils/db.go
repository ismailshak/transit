// Package testutils provides helper functions for testing purposes
//
// Should not be used by any non _test packages
package testutils

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/store"
)

// BlankDB creates a temporary database that has no tables
func BlankDB(t *testing.T) *store.Store {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-blank.db")

	t.Logf("Temp database at: %s", dbPath)

	db, err := store.NewStore(dbPath, logger.New(false))
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

// MigratedDB creates a temporary test database that's fully migrated
func MigratedDB(t *testing.T) *store.Store {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-migrated.db")

	db, err := store.NewStore(dbPath, logger.New(false))
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

func InitMigrationsTable(t *testing.T, db *sql.DB) {
	t.Helper()

	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %s", err)
	}

	defer tx.Rollback()

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
