package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/tui"
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
		ExecuteInitConfig()

		location := config.GetConfig().Core.Location
		client := api.GetClient(data.LocationSlug(location))

		if client == nil {
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		ExecuteInitData(client, data.LocationSlug(location))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func ExecuteInitConfig() {
	db, err := data.GetDB()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to database: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	locations, err := db.GetAllLocations()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get available locations: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	location := config.GetConfig().Core.Location
	if location == "" {
		selection := tui.NewListPrompt("Select a location", tui.ToListItems(locations)).Render()
		if selection == "" {
			tui.OperationSkipped("Canceled... Exiting")
			utils.Exit(utils.EXIT_SUCCESS)
		}

		err = ExecuteSet("core.location", selection)
		if err != nil {
			logger.Error(err)
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		location = selection
	}

	tui.OperationSuccessful("Location set to " + location)

	keyPath := fmt.Sprintf("%s.api_key", location)
	apiKey := ExecuteGet(keyPath)
	if apiKey == "" {
		key := tui.NewPasswordPrompt(fmt.Sprintf("Enter your API key for %s", location)).Render()
		if key == "" {
			tui.OperationSkipped("Canceled... Exiting")
			utils.Exit(utils.EXIT_SUCCESS)
		}

		err = ExecuteSet(keyPath, key)
		if err != nil {
			logger.Error(err)
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}
	}

	tui.OperationSuccessful("API key set")
}

func ExecuteInitData(client api.Api, location data.LocationSlug) {
	db, err := data.GetDB()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to database: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	count, err := db.CountStopsByLocation(location)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get status of current location: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	if count > 0 {
		tui.OperationSuccessful("Data initialized")
		return
	}

	fetchSpinner := tui.NewSpinner("Fetching data...")
	go fetchSpinner.Start()

	d, err := client.FetchStaticData()
	if err != nil {
		fetchSpinner.Stop()
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	fetchSpinner.Success("Data fetched")

	insertSpinner := tui.NewSpinner("Saving data...")
	go insertSpinner.Start()

	if err = db.InsertAgencies(d.Agencies); err != nil {
		insertSpinner.Stop()
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	if err = db.InsertStops(d.Stops); err != nil {
		insertSpinner.Stop()
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	insertSpinner.Success("Data saved")

	logger.Print("\nAll done. Use transit --help for commands and examples")
}
