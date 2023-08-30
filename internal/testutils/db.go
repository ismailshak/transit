// Package testutils provides helper functions for testing purposes
//
// Should not be used by any non _test packages
package testutils

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/data"
)

// Creates a temporary database that has no tables
func BlankDB(t *testing.T) *sql.DB {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "db_test.db")

	t.Logf("Temp database at: %s", dbPath)

	db, err := data.DbConnect(dbPath)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// Creates a temporary test database that's fully migrated
func MigratedDB(t *testing.T) *sql.DB {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "models_test.db")

	db, err := data.DbConnect(dbPath)
	if err != nil {
		t.Fatal("Failed to connect to test database", err)
	}

	if err = data.SyncMigrations(db); err != nil {
		t.Fatal("Failed to migrate test database", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}
