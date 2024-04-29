package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize transit",
	Long: `
Adds missing config properties and downloads static data for the chosen location`,
	Args:   cobra.NoArgs,
	PreRun: defaultPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		location := config.GetConfig().Core.Location
		client := api.GetClient(data.LocationSlug(location))

		if client == nil {
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		ExecuteInit(client, data.LocationSlug(location))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func ExecuteInit(client api.Api, location data.LocationSlug) {
	db, err := data.GetDB()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	count, err := db.CountStopsByLocation(location)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get status of current location: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	if count > 0 {
		logger.Print("Location already initialized")
		return
	}

	// TODO:
	// Wrap in goroutine and show spinner ":spinner: Fetching data" -> ":check: Fetching data"
	d, err := client.FetchStaticData()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to fetch data: %s", err))
		return
	}

	// TODO:
	// Wrap in a goroutine and show spinner ":spinner: Saving data" -> ":check: Saving data"
	if err = db.InsertAgencies(d.Agencies); err != nil {
		logger.Error(fmt.Sprintf("Failed to save agencies: %s", err))
		return
	}
	if err = db.InsertStops(d.Stops); err != nil {
		logger.Error(fmt.Sprintf("Failed to save stops: %s", err))
		return
	}

	logger.Print("All done. Use transit --help for list of commands")
}
