package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
	"github.com/spf13/cobra"
)

// App is the assembled program, holding everything a command needs. It's built
// once in Execute and never leaves this package.
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

func (a *App) configSetupPreRun() {
	cfg, err := config.Load(a.configOverride)
	if err != nil {
		a.Log.Error("Failed to load config: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	a.Cfg = cfg
}

func (a *App) dbSetupPreRun() {
	path, err := config.GetConfigDir()
	if err != nil {
		a.Log.Error("Failed to load store: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	db, err := data.NewDB(filepath.Join(path, "transit.db"), a.Log)
	if err != nil {
		a.Log.Error(err.Error())
		utils.Exit(utils.EXIT_FAILURE)
	}

	a.Store = db

	err = a.Store.SyncMigrations()
	if err != nil {
		a.Log.Error("Database sync failed: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}

func (a *App) defaultPreRun(cmd *cobra.Command, args []string) {
	a.configSetupPreRun()
	a.dbSetupPreRun()
}

// client returns the API client for the configured location.
func (a *App) client() (api.Api, error) {
	switch data.LocationSlug(a.Cfg.Core.Location) {
	case data.DMVSlug:
		if a.Cfg.DMV.ApiKey == "" {
			return nil, errors.New("no api key found in config at 'dmv.api_key'")
		}
		return api.NewDMV(a.Cfg.DMV.ApiKey, a.Store, a.Log), nil
	case data.SFSlug:
		if a.Cfg.SF.ApiKey == "" {
			return nil, errors.New("no api key found in config at 'sf.api_key'")
		}
		return api.NewSF(a.Cfg.SF.ApiKey, a.Store, a.Log, a.Now), nil
	default:
		return nil, fmt.Errorf("invalid location %q", a.Cfg.Core.Location)
	}
}
