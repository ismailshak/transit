package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/ismailshak/transit/internal/tui"
	"github.com/spf13/cobra"
)

func (a *App) newIncidentsCmd() *cobra.Command {
	incidentsCmd := &cobra.Command{
		Use:     "incidents",
		Aliases: []string{"inc"},
		Short:   "Display reported disruptions or delays",
		Args:    usageArgs(cobra.NoArgs),
		PreRunE: a.defaultPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}

			return a.executeIncidents(cmd.Context(), client)
		},
	}

	return incidentsCmd
}

func (a *App) executeIncidents(ctx context.Context, client provider.API) error {
	alertSet, err := client.FetchIncidents(ctx)
	if err != nil {
		return fmt.Errorf("fetch incidents: %w", err)
	}

	degraded := alertSet.Degraded()
	if len(alertSet.Sources) == 0 {
		return errors.New("no sources were asked for alerts")
	}

	if len(degraded) == len(alertSet.Sources) {
		return errors.New("every source failed")
	}

	agencies, err := a.Store.Agencies(ctx, transit.LocationSlug(a.Cfg.Core.Location))
	if err != nil {
		return fmt.Errorf("look up agencies: %w", err)
	}

	tui.PrintIncidents(client, alertSet, len(agencies) > 1)

	for _, s := range degraded {
		a.warnf("%v", s.Err)
	}

	return nil
}
