package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/ismailshak/transit/internal/tui"
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

func (a *App) executeAt(ctx context.Context, client provider.API, args []string) error {
	targets, err := a.resolveStops(ctx, client, args)
	if err != nil {
		return err
	}

	return a.renderDepartures(ctx, client, targets)
}

func (a *App) watchAt(ctx context.Context, client provider.API, args []string) error {
	interval, err := watchInterval(a.Cfg.Core.WatchInterval)
	if err != nil {
		return err
	}

	targets, err := a.resolveStops(ctx, client, args)
	if err != nil {
		return err
	}

	message := tui.Bold(fmt.Sprintf("Refreshing station arrivals every %v. Press Ctrl+C to quit.", interval))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	buffer := tui.NewBuffer()
	buffer.StartAlternateBuffer()
	defer buffer.StopAlternateBuffer()

	for {
		buffer.RefreshScreen()
		_, _ = fmt.Fprintln(a.Out, message)

		if err := a.renderDepartures(ctx, client, targets); err != nil {
			if endsWatch(err) {
				return err
			}

			if !errors.Is(err, provider.ErrNoDepartures) {
				a.errorf("%s", err)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// target is one argument resolved to the stop codes a provider wants
type target struct {
	arg  string
	refs []transit.StopRef
}

func (a *App) resolveStops(ctx context.Context, client provider.API, args []string) ([]target, error) {
	slug := transit.LocationSlug(a.Cfg.Core.Location)

	var targets []target
	for _, arg := range args {
		stops, err := a.Store.MatchStops(ctx, slug, arg)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", arg, err)
		}

		// TODO: report these to the caller so it can warn on nothing matching
		if len(stops) == 0 {
			continue
		}

		if len(stops) > 5 {
			continue
		}

		var refs []transit.StopRef
		for _, s := range stops {
			refs = append(refs, client.StopRefs(s)...)
		}

		targets = append(targets, target{arg: arg, refs: refs})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("%w: no station matched %v", errUsage, args)
	}

	return targets, nil
}

func (a *App) renderDepartures(ctx context.Context, client provider.API, targets []target) error {
	var rendered int
	for _, t := range targets {
		departures, err := client.FetchPredictions(ctx, t.refs)
		if errors.Is(err, provider.ErrNoDepartures) {
			continue
		}

		if err != nil {
			return fmt.Errorf("fetch predictions for %q: %w", t.arg, err)
		}

		destinationLookup, sortedDestinations := groupByDestination(departures)
		tui.PrintArrivalScreen(&destinationLookup, sortedDestinations, a.Now())
		rendered++
	}

	if rendered == 0 {
		return provider.ErrNoDepartures
	}

	return nil
}

// Groups predictions by destination (assumes already sorted by minutes).
// Sometimes the same destination can have multiple lines, so we group by both.
// Returns grouped map and returns a sorted list of destinations
func groupByDestination(departures []transit.Departure) (map[string][]transit.Departure, []string) {
	destMap := make(map[string][]transit.Departure)
	var destinations []string

	for _, d := range departures {
		key := fmt.Sprintf("%s-%s", d.Headsign, d.Line)
		_, exists := destMap[key]
		if exists {
			destMap[key] = append(destMap[key], d)
		} else {
			destMap[key] = []transit.Departure{d}
			destinations = append(destinations, key)
		}
	}

	sort.Strings(destinations)

	return destMap, destinations
}

// watchInterval converts the configured seconds into a duration. A non-positive
// value would panic time.NewTicker, and config set already refuses one.
func watchInterval(seconds int) (time.Duration, error) {
	if seconds <= 0 {
		return 0, fmt.Errorf("%w: watch_interval must be greater than 0", config.ErrInvalid)
	}
	return time.Duration(seconds) * time.Second, nil
}

// endsWatch reports whether an error should end the session. Only recoverable
// errors keep the loop going.
func endsWatch(err error) bool {
	// Check cancelled first as a cancelled request can surface as net.OpError below.
	if cancelled(err) {
		return true
	}

	if httpErr, ok := errors.AsType[*provider.HTTPError](err); ok {
		// 5xx and 429 are the upstream's problem and it could recover.
		return httpErr.StatusCode < 500 && httpErr.StatusCode != http.StatusTooManyRequests
	}

	// Things the network can fix on its own. Timeout arrives as DeadlineExceeded.
	if _, ok := errors.AsType[*net.OpError](err); ok {
		return false
	}

	// A name that never resolved didn't get as far as a dial so there's no OpError here.
	if _, ok := errors.AsType[*net.DNSError](err); ok {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Loop could start before trains start operating.
	if errors.Is(err, provider.ErrNoDepartures) {
		return false
	}

	return true
}
