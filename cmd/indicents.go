package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/logger"
	"github.com/ismailshak/transit/tui"
	"github.com/spf13/cobra"
)

var incidentsCmd = &cobra.Command{
	Use:     "incidents",
	Aliases: []string{"inc"},
	Short:   "Display reported disruptions or delays",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		location := config.GetConfig().Core.Location
		client := api.GetClient(location)
		ExecuteIncidents(client)
	},
}

func init() {
	rootCmd.AddCommand(incidentsCmd)
}

func ExecuteIncidents(client api.Api) {
	incidents, err := client.FetchIncidents()
	if err != nil {
		logger.Error(fmt.Sprint(err))
	}

	tui.PrintIssues(client, incidents)
}
