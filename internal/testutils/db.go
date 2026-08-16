// Package testutils provides helper functions for testing purposes
//
// Should not be used by any non _test packages
package testutils

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
)

// BlankDB creates a temporary database that has no tables
func BlankDB(t *testing.T) *data.TransitDB {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-blank.db")

	t.Logf("Temp database at: %s", dbPath)

	db, err := data.NewDB(dbPath, logger.New(false))
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
func MigratedDB(t *testing.T) *data.TransitDB {
	t.Helper()
	testDir := t.TempDir()
	dbPath := filepath.Join(testDir, "transit-test-migrated.db")

	db, err := data.NewDB(dbPath, logger.New(false))
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

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %s", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(data.CreateMigrationsTableSQL)
	if err != nil {
		t.Fatalf("Failed to create migration: %s", err)
	}

	_, err = tx.Exec(data.InsertMigrationSQL, "1_FakeMigration")

	tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %s", err)
	}
}
