package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/internal/ui"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

func (a *App) newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize transit",
		Long: `
Adds missing config properties and downloads static data for the chosen location`,
		Args:   cobra.NoArgs,
		PreRun: a.defaultPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			a.executeInitConfig(cmd.Context())

			client, err := a.client()
			if err != nil {
				a.Log.Error(err.Error())
				utils.Exit(utils.EXIT_BAD_CONFIG)
			}

			a.executeInitData(cmd.Context(), client, data.LocationSlug(a.Cfg.Core.Location))
		},
	}

	return initCmd
}

func toChoices(locations []data.Location) []ui.Choice {
	choices := make([]ui.Choice, len(locations))
	for i, l := range locations {
		choices[i] = ui.Choice{Key: string(l.Slug), Title: string(l.Slug), Description: l.Name, FilterValue: l.Name}
	}

	return choices
}

func (a *App) getConfiguredLocation(ctx context.Context) string {
	location := a.Cfg.Core.Location
	if location != "" {
		return location
	}

	locations, err := a.Store.GetAllLocations()
	if err != nil {
		a.Log.Error(fmt.Sprintf("failed to get locations: %s", err))
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
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_FAILURE)
	}

	err = a.executeSet("core.location", selection)
	if err != nil {
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	return selection
}

func (a *App) confirmConfiguredKey(ctx context.Context, location string) {
	keyPath := fmt.Sprintf("%s.api_key", location)
	apiKey := a.executeGet(keyPath)
	if apiKey != "" {
		return
	}

	key, err := ui.Password(ctx, fmt.Sprintf("Enter your API key for %s", location))

	if err != nil {
		if errors.Is(err, ui.ErrCancelled) {
			tui.OperationSkipped("Cancelled... Exiting")
			utils.Exit(utils.EXIT_SUCCESS)
		}

		if errors.Is(err, ui.ErrNoInput) {
			tui.OperationFailed("No input... Exiting")
			utils.Exit(utils.EXIT_BAD_USAGE)
		}

		tui.OperationFailed("Failed to capture input")
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_FAILURE)
	}

	err = a.executeSet(keyPath, key)
	if err != nil {
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}

func (a *App) executeInitConfig(ctx context.Context) {
	location := a.getConfiguredLocation(ctx)
	tui.OperationSuccessful("Location set to " + location)
	a.confirmConfiguredKey(ctx, location)
	tui.OperationSuccessful("API key set")
}

func (a *App) executeInitData(ctx context.Context, client api.Api, location data.LocationSlug) {
	count, err := a.Store.CountStopsByLocation(location)
	if err != nil {
		a.Log.Error(fmt.Sprintf("failed to get status of current location: %s", err))
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
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_FAILURE)
	}

	err = ui.WithSpinner(ctx, &ui.SpinnerOptions{
		SpinMessage:    "Saving data...",
		ErrorMessage:   "Failed to save data",
		SuccessMessage: "Data saved",
		CallbackFn: func() error {
			if insertErr := a.Store.InsertAgencies(d.Agencies); insertErr != nil {
				return insertErr
			}

			if insertErr := a.Store.InsertStops(d.Stops); insertErr != nil {
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
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_FAILURE)
	}

	_, _ = fmt.Fprintln(a.Out, "\nSuccessfully initialized. Use transit --help for commands and examples")
}
