package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
	"github.com/ismailshak/transit/tui"
	"github.com/spf13/cobra"
)

var atCmd = &cobra.Command{
	Use:     "at <args>",
	Example: "  transit at courth (matches \"Court House\")\n  transit at metro (matches \"Metro Center\")",
	Short:   "Display upcoming train arrival information at chosen station(s)",
	Long: `
Display upcoming train information for one or more stations.

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

		ExecuteAt(client, args)
	},
}

func init() {
	rootCmd.AddCommand(atCmd)
}

// Entry point to the `at` subcommand
func ExecuteAt(client api.Api, args []string) {
	for _, arg := range args {
		codes := client.GetCodeFromArg(arg)
		if codes == nil {
			continue
		}

		predictions, err := client.FetchPredictions(codes)
		if err != nil {
			logger.Error(fmt.Sprint(err))
			helpers.Exit(helpers.EXIT_BAD_CONFIG)
		}

		tui.PrintArrivingScreen(predictions)
	}
}
