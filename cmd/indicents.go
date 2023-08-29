package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/pkg/api"
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
