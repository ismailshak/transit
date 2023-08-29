package data_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/data"
)

func getTestDb(t *testing.T) *sql.DB {
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

func TestMigrationCompletes(t *testing.T) {
	db := getTestDb(t)
	err := data.SyncMigrations(db)

	if err != nil {
		t.Errorf("Failed to sync migrations. %s", err)
	}
}
