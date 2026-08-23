package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/render"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/ismailshak/transit/internal/tui"
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

			p, err := a.provider()
			if err != nil {
				return err
			}

			seeder, ok := p.(transit.Seeder)
			if !ok {
				return nil
			}

			if err := a.executeInitData(ctx, seeder, transit.LocationSlug(a.Cfg.Core.Location)); err != nil {
				return fmt.Errorf("initialize data: %w", err)
			}

			return nil
		},
	}

	return initCmd
}

func toChoices(locations []transit.Location) []tui.Choice {
	choices := make([]tui.Choice, len(locations))
	for i, l := range locations {
		choices[i] = tui.Choice{Key: string(l.Slug), Title: string(l.Slug), Description: l.Name, FilterValue: l.Name}
	}

	return choices
}

func (a *App) getConfiguredLocation(ctx context.Context) (string, error) {
	location := a.Cfg.Core.Location
	if location != "" {
		return location, nil
	}

	locations, err := a.Store.AllLocations(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch locations: %w", err)
	}

	choices := toChoices(locations)

	selection, err := tui.Select(ctx, "Select a location", choices)
	if err != nil {
		if errors.Is(err, tui.ErrCancelled) {
			a.print(render.Skipped("Cancelled... Exiting"))
			return "", tui.ErrCancelled
		}

		if errors.Is(err, tui.ErrNoSelection) {
			a.print(render.Skipped("Nothing selected... Exiting"))
			return "", tui.ErrNoSelection
		}

		a.print(render.Failed("Failed to select location"))
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

	key, err := tui.Password(ctx, fmt.Sprintf("Enter your API key for %s", location))
	if err != nil {
		if errors.Is(err, tui.ErrCancelled) {
			a.print(render.Skipped("Cancelled... Exiting"))
			return err
		}

		if errors.Is(err, tui.ErrNoInput) {
			a.print(render.Failed("No input... Exiting"))
			return err
		}

		a.print(render.Failed("Failed to capture input"))
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

	a.print(render.Success("Location set to " + location))

	err = a.confirmConfiguredKey(ctx, location)
	if err != nil {
		return err
	}

	a.print(render.Success("API key set"))
	return nil
}

func (a *App) executeInitData(ctx context.Context, seeder transit.Seeder, location transit.LocationSlug) error {
	count, err := a.Store.CountStopsByLocation(ctx, location)
	if err != nil {
		return fmt.Errorf("count stops: %w", err)
	}

	if count > 0 {
		a.print(render.Success("Data initialized"))
		return nil
	}

	var d *transit.Static
	err = tui.WithSpinner(ctx, &tui.SpinnerOptions{
		SpinMessage: "Fetching data...",
		CallbackFn: func(ctx context.Context) error {
			var fetchErr error
			d, fetchErr = seeder.Seed(ctx)
			return fetchErr
		},
	})

	if errors.Is(err, tui.ErrCancelled) {
		a.print(render.Skipped("Cancelled... Exiting"))
		return tui.ErrCancelled
	}

	if err != nil {
		a.print(render.Failed("Failed to fetch data"))
		return fmt.Errorf("fetch static data: %w", err)
	}

	a.print(render.Success("Data fetched"))

	err = tui.WithSpinner(ctx, &tui.SpinnerOptions{
		SpinMessage: "Saving data...",
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

	if errors.Is(err, tui.ErrCancelled) {
		a.print(render.Skipped("Cancelled... Exiting"))
		return tui.ErrCancelled
	}

	if err != nil {
		a.print(render.Failed("Failed to save data"))
		return fmt.Errorf("insert data: %w", err)
	}

	a.print(render.Success("Data saved"))
	a.print("\nSuccessfully initialized. Use transit --help for commands and examples")

	return nil
}
