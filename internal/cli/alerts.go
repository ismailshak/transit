package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/ismailshak/transit/internal/render"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/spf13/cobra"
)

func (a *App) newAlertsCmd() *cobra.Command {
	alertsCmd := &cobra.Command{
		Use:     "alerts",
		Short:   "Display reported disruptions or delays",
		Args:    usageArgs(cobra.NoArgs),
		PreRunE: a.defaultPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := a.provider()
			if err != nil {
				return err
			}

			return a.executeAlerts(cmd.Context(), p)
		},
	}

	return alertsCmd
}

func (a *App) executeAlerts(ctx context.Context, p transit.Provider) error {
	alertSet, err := p.Alerts(ctx)
	if err != nil {
		return fmt.Errorf("fetch alerts: %w", err)
	}

	degraded := alertSet.Degraded()
	if len(alertSet.Alerts) == 0 && len(degraded) > 0 {
		return fmt.Errorf("fetch alerts: %w", errors.Join(errsOf(degraded)...))
	}

	agencies, err := a.Store.Agencies(ctx, transit.LocationSlug(a.Cfg.Core.Location))
	if err != nil {
		return fmt.Errorf("look up agencies: %w", err)
	}

	a.println(render.Alerts{Set: alertSet, Width: a.Out.Width(), ShowAgency: len(agencies) > 1})

	for _, s := range degraded {
		a.warnf("%v", s.Err)
	}

	return nil
}

func errsOf(sources []transit.SourceStatus) []error {
	errs := make([]error, 0, len(sources))
	for _, s := range sources {
		errs = append(errs, s.Err)
	}

	return errs
}
