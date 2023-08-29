package data

import (
	"database/sql"
	"path/filepath"

	"github.com/ismailshak/transit/internal/utils"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func GetDBConn() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	configPath := utils.GetConfigDir()
	dbPath := filepath.Join(configPath, "transit.db")
	newDb, err := DbConnect(dbPath)

	if err != nil {
		return nil, err
	}

	db = newDb

	return db, nil
}

// Keep migrations up-to-date, and handle first time migration run
func SyncMigrations(db *sql.DB) error {
	err := createMigrationTable(db)
	if err != nil {
		return err
	}

	count, err := getMigrationCount(db)
	if err != nil {
		return err
	}

	if count == len(migrationChangesets) {
		return nil
	}

	err = runMigrations(db, count)
	if err != nil {
		return err
	}

	return nil
}

// Exists for testing purposes. Use GetDBConn instead
func DbConnect(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}
