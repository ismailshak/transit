package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/ismailshak/transit/internal/tui"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

func (a *App) newAtCmd() *cobra.Command {
	var watchFlag bool

	atCmd := &cobra.Command{
		Use:     "at <args>",
		Example: "  transit at courth (matches \"Court House\")\n  transit at metro (matches \"Metro Center\")",
		Short:   "Display upcoming train arrival information at chosen station(s)",
		Long: `
Display upcoming train information for one or more stations.

Arguments are considered valid if it can be used to narrow
the official station names to just 1. If something's too generic,
try being more specific by adding more characters.
	`,
		Args:    usageArgs(cobra.MinimumNArgs(1)),
		PreRunE: a.defaultPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := a.client()
			if err != nil {
				return err
			}

			if watchFlag {
				return a.watchAt(client, args)
			}

			return a.executeAt(client, args)
		},
	}

	atCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "live update arrival information")

	return atCmd
}

func (a *App) executeAt(client api.Api, args []string) error {
	var rendered, resolved int

	// TODO: pull client.GetIDFromArg() out of this so that `Watch` is more performant
	for _, arg := range args {
		codes, err := client.GetPredictionInput(arg)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", arg, err)
		}
		if codes == nil {
			continue
		}

		resolved++

		predictions, err := client.FetchPredictions(codes)
		if errors.Is(err, api.ErrNoDepartures) {
			continue
		}

		if err != nil {
			return fmt.Errorf("fetch predictions for %q: %w", arg, err)
		}

		destinationLookup, sortedDestinations := groupByDestination(predictions)
		tui.PrintArrivalScreen(client, &destinationLookup, sortedDestinations)
		rendered++
	}

	if resolved == 0 {
		return fmt.Errorf("%w: no station matched %v", errUsage, args)
	}

	if rendered == 0 {
		return api.ErrNoDepartures
	}

	return nil
}

func (a *App) watchAt(client api.Api, args []string) error {
	buffer := tui.NewBuffer()
	interval := time.Second * time.Duration(a.Cfg.Core.WatchInterval)
	message := tui.Bold(fmt.Sprintf("Refreshing station arrivals every %v. Press Ctrl+C to quit.", interval))
	cancelChan := make(chan os.Signal, 1)

	// catch SIGTERM or SIGINT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	buffer.StartAlternateBuffer()

	go func() {
		for {
			buffer.RefreshScreen()
			_, _ = fmt.Fprintln(a.Out, message)
			// Watch mode reports errors on screen rather than ending the loop
			if err := a.executeAt(client, args); err != nil && !errors.Is(err, api.ErrNoDepartures) {
				a.Log.Error(err.Error())
			}

			time.Sleep(interval)
		}
	}()

	// blocking expression
	<-cancelChan

	buffer.StopAlternateBuffer()
	return nil
}

// Groups predictions by destination (assumes already sorted by minutes).
// Sometimes the same destination can have multiple lines, so we group by both.
// Returns grouped map and returns a sorted list of destinations
func groupByDestination(predictions []api.Prediction) (map[string][]api.Prediction, []string) {
	destMap := make(map[string][]api.Prediction)
	var destinations []string

	for _, p := range predictions {
		key := fmt.Sprintf("%s-%s", p.Destination, p.Line)
		_, exists := destMap[key]
		if exists {
			destMap[key] = append(destMap[key], p)
		} else {
			destMap[key] = []api.Prediction{p}
			destinations = append(destinations, key)
		}
	}

	sort.Strings(destinations)

	return destMap, destinations
}
