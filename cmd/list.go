package cmd

import (
	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/list"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list <args>",
	Example: "  transit list courth (matches \"Court House\")\n  transit list metro (matches \"Metro Center\")",
	Short:   "Display upcoming train arrival information for chosen station(s)",
	Long: `
'list' will display upcoming train information for one or more stations.

Arguments are considered valid if it can be used to narrow
the official station names to just 1. If something's too generic,
try being more specific by adding more characters.
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		location := config.GetConfig().Core.Location
		client := api.GetClient(location)
		if client == nil {
			helpers.Exit(helpers.EXIT_BAD_CONFIG)
		}

		list.Execute(client, args)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
