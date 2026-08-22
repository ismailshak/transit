package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/spf13/cobra"
)

// App is the assembled program, holding everything a command needs. It's built
// once in [Run] and should not be used in other packages.
type App struct {
	Cfg   *config.Config
	Store *store.Store
	Out   io.Writer
	Err   io.Writer
	Now   func() time.Time

	// Bound to --config in newRootCmd.
	// Empty means the default location.
	configOverride string

	// Bound to --verbose in newRootCmd.
	verbose bool
}

// run executes the command tree against args and returns a process exit code.
// Args are passed in rather than read from os.Args to make testing easier.
func (a *App) run(ctx context.Context, args []string) int {
	cmd := a.newRootCmd()
	cmd.SetArgs(args)

	err := cmd.ExecuteContext(ctx)
	if err != nil && !cancelled(err) {
		a.errorf("%s", err)
	}

	return exitCode(err)
}

// close releases anything a hook opened. Commands that never initialize
// the store leave it nil, so it has to handle that.
func (a *App) close() error {
	if a.Store == nil {
		return nil
	}

	return a.Store.Close()
}

// configSetupPreRun is the hook for commands that only read the config file.
func (a *App) configSetupPreRun(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load(a.configOverride)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	a.Cfg = cfg
	return nil
}

func (a *App) dbSetupPreRun(ctx context.Context) error {
	path, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("locate config: %w", err)
	}

	db, err := store.New(filepath.Join(path, "transit.db"))
	if err != nil {
		return fmt.Errorf("establish store: %w", err)
	}

	a.Store = db

	err = a.Store.SyncMigrations(ctx)
	if err != nil {
		return fmt.Errorf("synchronize migrations: %w", err)
	}

	return nil
}

// defaultPreRun is the hook for commands that read the config and the store.
func (a *App) defaultPreRun(cmd *cobra.Command, args []string) error {
	err := a.configSetupPreRun(cmd, args)
	if err != nil {
		return err
	}

	return a.dbSetupPreRun(cmd.Context())
}

// provider returns the configured location's provider.
func (a *App) provider() (transit.Provider, error) {
	switch transit.LocationSlug(a.Cfg.Core.Location) {
	case transit.DMVSlug:
		client, err := provider.NewDMV(a.Cfg.DMV.APIKey, a.Now)
		if err != nil {
			return nil, fmt.Errorf("dmv client: %w", err)
		}
		return client, nil
	case transit.SFSlug:
		client, err := provider.NewSF(a.Cfg.SF.APIKey, a.Store)
		if err != nil {
			return nil, fmt.Errorf("sf client: %w", err)
		}
		return client, nil
	default:
		return nil, fmt.Errorf("%w: unsupported location %q", config.ErrInvalid, a.Cfg.Core.Location)
	}
}
