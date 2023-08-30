// Package data implements functions that interact with all pieces of
// data used by transit CLI.
//
// This can be data stored in the SQLite database on a user's machine, or
// data downloaded from a server like GTFS
package data

import (
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/spf13/cobra"
)

// Static data, that doesn't change often, that we store in the database
type StaticData struct {
	Stops []*Stop
}

// PreRun helper that makes sure the database is up-to-date
func CommandPre(cmd *cobra.Command, args []string) {
	db, err := GetDBConn()
	if err != nil {
		logger.Error("Failed to connect to database: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	err = db.SyncMigrations()
	if err != nil {
		logger.Error("Database sync failed: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}
