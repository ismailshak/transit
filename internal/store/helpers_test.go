package store_test

import (
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

	db, err := store.New(dbPath)
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

	db, err := store.New(dbPath)
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
