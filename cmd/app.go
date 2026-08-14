package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

// App is the assembled program, holding everything a command needs. It's built
// once in [Run] and never leaves this package.
type App struct {
	Cfg   *config.Config
	Store *data.TransitDB
	Log   *logger.Logger
	Out   io.Writer
	Now   func() time.Time

	// Bound to --config in newRootCmd, read in configSetupPreRun.
	// Empty means the default location.
	configOverride string
}

func (a *App) configSetupPreRun() error {
	cfg, err := config.Load(a.configOverride)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	a.Cfg = cfg
	return nil
}

func (a *App) dbSetupPreRun() error {
	path, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("locate config: %w", err)
	}

	db, err := data.NewDB(filepath.Join(path, "transit.db"), a.Log)
	if err != nil {
		return fmt.Errorf("establish store: %w", err)
	}

	a.Store = db

	err = a.Store.SyncMigrations()
	if err != nil {
		return fmt.Errorf("synchronize migrations: %w", err)
	}

	return nil
}

func (a *App) defaultPreRun(cmd *cobra.Command, args []string) error {
	err := a.configSetupPreRun()
	if err != nil {
		return err
	}

	return a.dbSetupPreRun()
}

// client returns the API client for the configured location.
func (a *App) client() (api.Api, error) {
	switch data.LocationSlug(a.Cfg.Core.Location) {
	case data.DMVSlug:
		client, err := api.NewDMV(a.Cfg.DMV.ApiKey, a.Store, a.Log)
		if err != nil {
			return nil, fmt.Errorf("dmv client: %w", err)
		}
		return client, nil
	case data.SFSlug:
		client, err := api.NewSF(a.Cfg.SF.ApiKey, a.Store, a.Log, a.Now)
		if err != nil {
			return nil, fmt.Errorf("sf client: %w", err)
		}
		return client, nil
	default:
		return nil, fmt.Errorf("%w: unsupported location %q", config.ErrInvalid, a.Cfg.Core.Location)
	}
}
