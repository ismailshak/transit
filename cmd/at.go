package cmd

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

			ctx := cmd.Context()

			if watchFlag {
				return a.watchAt(ctx, client, args)
			}

			return a.executeAt(ctx, client, args)
		},
	}

	atCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "live update arrival information")

	return atCmd
}

func (a *App) executeAt(ctx context.Context, client api.API, args []string) error {
	var rendered, resolved int

	// TODO: pull client.GetPredictionInput() out of this so that `Watch` is more performant
	for _, arg := range args {
		codes, err := client.GetPredictionInput(ctx, arg)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", arg, err)
		}
		if codes == nil {
			continue
		}

		resolved++

		predictions, err := client.FetchPredictions(ctx, codes)
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

func (a *App) watchAt(ctx context.Context, client api.API, args []string) error {
	buffer := tui.NewBuffer()
	interval := time.Second * time.Duration(a.Cfg.Core.WatchInterval)
	message := tui.Bold(fmt.Sprintf("Refreshing station arrivals every %v. Press Ctrl+C to quit.", interval))

	buffer.StartAlternateBuffer()

	go func() {
		for {
			buffer.RefreshScreen()
			_, _ = fmt.Fprintln(a.Out, message)
			// Watch mode reports errors on screen rather than ending the loop
			if err := a.executeAt(ctx, client, args); err != nil && !errors.Is(err, api.ErrNoDepartures) && !cancelled(err) {
				a.Log.Error(err.Error())
			}

			time.Sleep(interval)
		}
	}()

	// blocking expression
	<-ctx.Done()

	buffer.StopAlternateBuffer()

	// Return cancelled context so that the correct exit code is returned
	return ctx.Err()
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
