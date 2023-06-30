package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
	"github.com/ismailshak/transit/tui"
	"github.com/spf13/cobra"
)

// Used for flags
var (
	watchFlag bool
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

		if watchFlag {
			WatchExecuteAt(client, args)
			return
		}

		ExecuteAt(client, args)
	},
}

func init() {
	rootCmd.AddCommand(atCmd)

	atCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "live update arrival information")
}

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

		destinationLookup, sortedDestinations := groupByDestination(predictions)
		tui.PrintArrivingScreen(client, &destinationLookup, sortedDestinations)
	}
}

func WatchExecuteAt(client api.Api, args []string) {
	buffer := tui.NewBuffer()
	interval := time.Second * time.Duration(config.GetConfig().Core.WatchInterval)
	message := tui.Bold(fmt.Sprintf("Refreshing station arrivals every %v. Press Ctrl+C to quit.", interval))
	cancelChan := make(chan os.Signal, 1)

	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	buffer.StartAlternateBuffer()

	go func() {
		for {
			buffer.RefreshScreen()
			logger.Print(message)
			ExecuteAt(client, args)
			time.Sleep(interval)
		}
	}()

	// blocking expression
	<-cancelChan

	buffer.StopAlternateBuffer()
}

// Groups predictions by destination (assumes already sorted by minutes).
// Returns grouped map and returns a sorted list of destinations
func groupByDestination(predictions []api.Predictions) (map[string][]api.Predictions, []string) {
	destMap := make(map[string][]api.Predictions)
	var destinations []string

	for _, p := range predictions {
		_, exists := destMap[p.Destination]
		if exists {
			destMap[p.Destination] = append(destMap[p.Destination], p)
		} else {
			destMap[p.Destination] = []api.Predictions{p}
			destinations = append(destinations, p.Destination)
		}
	}

	sort.Strings(destinations)

	return destMap, destinations
}
