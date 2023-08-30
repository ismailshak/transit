package data

import (
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
	"github.com/spf13/cobra"
)

type Data struct {
	Stops []*Stop
}

// PreRun helper that makes sure the database is up-to-date
func CommandPre(cmd *cobra.Command, args []string) {
	dbConn, err := GetDBConn()
	if err != nil {
		logger.Error("Failed to connect to database: " + err.Error())
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	err = SyncMigrations(dbConn)
	if err != nil {
		logger.Error("Database sync failed: " + err.Error())
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}
}
