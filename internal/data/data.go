package data

import (
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
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
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	err = SyncMigrations(dbConn)
	if err != nil {
		logger.Error("Database sync failed: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}
