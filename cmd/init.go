package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/internal/ui"
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
		ExecuteInitConfig(cmd.Context())

		location := config.GetConfig().Core.Location
		client := api.GetClient(data.LocationSlug(location))

		if client == nil {
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		ExecuteInitData(cmd.Context(), client, data.LocationSlug(location))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func toChoices(locations []data.Location) []ui.Choice {
	choices := make([]ui.Choice, len(locations))
	for i, l := range locations {
		choices[i] = ui.Choice{Key: string(l.Slug), Title: string(l.Slug), Description: l.Name, FilterValue: l.Name}
	}

	return choices
}

func getConfiguredLocation(ctx context.Context) string {
	location := config.GetConfig().Core.Location
	if location != "" {
		return location
	}

	db, err := data.GetDB()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to database: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	locations, err := db.GetAllLocations()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get locations: %s", err))
		utils.Exit(utils.EXIT_BAD_USAGE) // TODO: replace error code with something database specific
	}

	choices := toChoices(locations)

	selection, err := ui.Select(ctx, "Select a location", choices)
	if err != nil {
		if errors.Is(err, ui.ErrCancelled) {
			tui.OperationSkipped("Cancelled... Exiting")
			utils.Exit(utils.EXIT_SUCCESS)
		}

		if errors.Is(err, ui.ErrNoSelection) {
			tui.OperationSkipped("Nothing selected... Exiting")
			utils.Exit(utils.EXIT_BAD_USAGE)
		}

		tui.OperationFailed("Failed to select location")
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	err = ExecuteSet("core.location", selection)
	if err != nil {
		logger.Error(err)
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	return selection
}

func confirmConfiguredKey(ctx context.Context, location string) {
	keyPath := fmt.Sprintf("%s.api_key", location)
	apiKey := ExecuteGet(keyPath)
	if apiKey != "" {
		return
	}

	key, err := ui.Password(ctx, fmt.Sprintf("Enter your API key for %s", location))

	if errors.Is(err, ui.ErrCancelled) {
		tui.OperationSkipped("Cancelled... Exiting")
		utils.Exit(utils.EXIT_SUCCESS)
	}

	if errors.Is(err, ui.ErrNoInput) {
		tui.OperationFailed("No input... Exiting")
		utils.Exit(utils.EXIT_BAD_USAGE)
	}

	err = ExecuteSet(keyPath, key)
	if err != nil {
		logger.Error(err)
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}

func ExecuteInitConfig(ctx context.Context) {
	location := getConfiguredLocation(ctx)
	tui.OperationSuccessful("Location set to " + location)
	confirmConfiguredKey(ctx, location)
	tui.OperationSuccessful("API key set")
}

func ExecuteInitData(ctx context.Context, client api.Api, location data.LocationSlug) {
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

	var d *data.StaticData
	err = ui.WithSpinner(ctx, &ui.SpinnerOptions{
		SpinMessage:    "Fetching data...",
		ErrorMessage:   "Failed to fetch data",
		SuccessMessage: "Data fetched",
		CallbackFn: func() error {
			var fetchErr error
			d, fetchErr = client.FetchStaticData()
			return fetchErr
		},
	})

	if errors.Is(err, ui.ErrCancelled) {
		tui.OperationSkipped("Cancelled... Exiting")
		utils.Exit(utils.EXIT_SUCCESS)
	}

	if err != nil {
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	err = ui.WithSpinner(ctx, &ui.SpinnerOptions{
		SpinMessage:    "Saving data...",
		ErrorMessage:   "Failed to save data",
		SuccessMessage: "Data saved",
		CallbackFn: func() error {
			if insertErr := db.InsertAgencies(d.Agencies); insertErr != nil {
				return insertErr
			}

			if insertErr := db.InsertStops(d.Stops); insertErr != nil {
				return insertErr
			}

			return nil
		},
	})

	if errors.Is(err, ui.ErrCancelled) {
		tui.OperationSkipped("Cancelled... Exiting")
		utils.Exit(utils.EXIT_SUCCESS)
	}

	if err != nil {
		logger.Error(err)
		utils.Exit(utils.EXIT_FAILURE)
	}

	logger.Print("\nSuccessfully initialized. Use transit --help for commands and examples")
}
