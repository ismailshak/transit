package data

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func GetDBConn() *sql.DB {
	if db != nil {
		return db
	}

	configPath := helpers.GetConfigDir()
	dbPath := filepath.Join(configPath, "transit.db")
	newDb, err := DbConnect(dbPath)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database. %s", err))
		helpers.Exit(1)
	}

	db = newDb

	return db
}

func DbConnect(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
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
