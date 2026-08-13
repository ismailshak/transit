package cmd

import (
	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

func (a *App) newIncidentsCmd() *cobra.Command {
	incidentsCmd := &cobra.Command{
		Use:     "incidents",
		Aliases: []string{"inc"},
		Short:   "Display reported disruptions or delays",
		Args:    cobra.NoArgs,
		PreRun:  a.defaultPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := a.client()
			if err != nil {
				a.Log.Error(err.Error())
				utils.Exit(utils.EXIT_BAD_CONFIG)
			}

			a.executeIncidents(client)
		},
	}

	return incidentsCmd
}

func (a *App) executeIncidents(client api.Api) {
	incidents, err := client.FetchIncidents()
	if err != nil {
		a.Log.Error(err.Error())
	}

	tui.PrintIncidents(client, incidents)
}
