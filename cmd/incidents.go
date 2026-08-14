package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/pkg/api"
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

			return a.executeIncidents(client)
		},
	}

	return incidentsCmd
}

func (a *App) executeIncidents(client api.Api) error {
	incidents, err := client.FetchIncidents()
	if err != nil {
		return fmt.Errorf("fetch incidents: %w", err)
	}

	tui.PrintIncidents(client, incidents)
	return nil
}
