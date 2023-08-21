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
	newDb, err := sql.Open("sqlite", dbPath)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database. %s", err))
		helpers.Exit(1)
	}

	db = newDb

	return db
}

// Keep migrations up-to-date, and handle first time migration run
func SyncMigrations(db *sql.DB) {
	createMigrationTable(db)
	latestMigrations(db)
}
