package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/internal/ui"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

func (a *App) newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize transit",
		Long: `
Adds missing config properties and downloads static data for the chosen location`,
		Args:    usageArgs(cobra.NoArgs),
		PreRunE: a.defaultPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := a.executeInitConfig(ctx); err != nil {
				return fmt.Errorf("collect information: %w", err)
			}

			client, err := a.client()
			if err != nil {
				return err
			}

			if err := a.executeInitData(ctx, client, data.LocationSlug(a.Cfg.Core.Location)); err != nil {
				return fmt.Errorf("initialize data: %w", err)
			}

			return nil
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

func (a *App) getConfiguredLocation(ctx context.Context) (string, error) {
	location := a.Cfg.Core.Location
	if location != "" {
		return location, nil
	}

	locations, err := a.Store.GetAllLocations(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch locations: %w", err)
	}

	choices := toChoices(locations)

	selection, err := ui.Select(ctx, "Select a location", choices)
	if err != nil {
		if errors.Is(err, ui.ErrCancelled) {
			tui.OperationSkipped("Cancelled... Exiting")
			return "", ui.ErrCancelled
		}

		if errors.Is(err, ui.ErrNoSelection) {
			tui.OperationSkipped("Nothing selected... Exiting")
			return "", ui.ErrNoSelection
		}

		tui.OperationFailed("Failed to select location")
		return "", err
	}

	err = a.executeSet(ctx, "core.location", selection)
	if err != nil {
		return "", fmt.Errorf("set location: %w", err)
	}

	return selection, nil
}

func (a *App) confirmConfiguredKey(ctx context.Context, location string) error {
	keyPath := fmt.Sprintf("%s.api_key", location)
	apiKey := a.executeGet(keyPath)
	if apiKey != "" {
		return nil
	}

	key, err := ui.Password(ctx, fmt.Sprintf("Enter your API key for %s", location))
	if err != nil {
		if errors.Is(err, ui.ErrCancelled) {
			tui.OperationSkipped("Cancelled... Exiting")
			return err
		}

		if errors.Is(err, ui.ErrNoInput) {
			tui.OperationFailed("No input... Exiting")
			return err
		}

		tui.OperationFailed("Failed to capture input")
		return err
	}

	err = a.executeSet(ctx, keyPath, key)
	if err != nil {
		return fmt.Errorf("set api key: %w", err)
	}

	return nil
}

func (a *App) executeInitConfig(ctx context.Context) error {
	location, err := a.getConfiguredLocation(ctx)
	if err != nil {
		return err
	}

	tui.OperationSuccessful("Location set to " + location)

	err = a.confirmConfiguredKey(ctx, location)
	if err != nil {
		return err
	}

	tui.OperationSuccessful("API key set")
	return nil
}

func (a *App) executeInitData(ctx context.Context, client api.Api, location data.LocationSlug) error {
	count, err := a.Store.CountStopsByLocation(ctx, location)
	if err != nil {
		return fmt.Errorf("count stops: %w", err)
	}

	if count > 0 {
		tui.OperationSuccessful("Data initialized")
		return nil
	}

	var d *data.StaticData
	err = ui.WithSpinner(ctx, &ui.SpinnerOptions{
		SpinMessage:    "Fetching data...",
		ErrorMessage:   "Failed to fetch data",
		SuccessMessage: "Data fetched",
		CallbackFn: func(ctx context.Context) error {
			var fetchErr error
			d, fetchErr = client.FetchStaticData(ctx)
			return fetchErr
		},
	})

	if errors.Is(err, ui.ErrCancelled) {
		tui.OperationSkipped("Cancelled... Exiting")
		return ui.ErrCancelled
	}

	if err != nil {
		return fmt.Errorf("fetch static data: %w", err)
	}

	err = ui.WithSpinner(ctx, &ui.SpinnerOptions{
		SpinMessage:    "Saving data...",
		ErrorMessage:   "Failed to save data",
		SuccessMessage: "Data saved",
		CallbackFn: func(ctx context.Context) error {
			if insertErr := a.Store.InsertAgencies(ctx, d.Agencies); insertErr != nil {
				return insertErr
			}

			if insertErr := a.Store.InsertStops(ctx, d.Stops); insertErr != nil {
				return insertErr
			}

			return nil
		},
	})

	if errors.Is(err, ui.ErrCancelled) {
		tui.OperationSkipped("Cancelled... Exiting")
		return ui.ErrCancelled
	}

	if err != nil {
		return fmt.Errorf("insert data: %w", err)
	}

	_, err = fmt.Fprintln(a.Out, "\nSuccessfully initialized. Use transit --help for commands and examples")
	return err
}
