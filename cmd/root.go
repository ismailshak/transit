package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/ui"
	"github.com/ismailshak/transit/internal/version"
	"github.com/spf13/cobra"
)

func (a *App) newRootCmd() *cobra.Command {
	var versionFlag bool // TODO; make -v the version flag
	var verboseFlag bool // TODO; turn this into a --level flag or keep as verbose?

	rootCmd := &cobra.Command{
		Use:           "transit",
		Short:         "Tool for interacting with local transit information",
		SilenceErrors: true,
		SilenceUsage:  true,
		// Root takes no positional args, so anything here is an unrecognised
		// command. Cobra would say the same thing implicitly via legacyArgs;
		// stating it lets the failure carry errUsage.
		Args: usageArgs(cobra.NoArgs),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			a.Log = logger.New(verboseFlag)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				version.Execute()
				return nil
			}

			if err := cmd.Help(); err != nil {
				return err
			}

			return nil
		},
	}

	// Inherited by every subcommand, so a bad flag is tagged wherever it's parsed.
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errUsage, err)
	})

	// Global, persistent flags
	rootCmd.PersistentFlags().StringVarP(&a.configOverride, "config", "c", "", "config file (defaults to $HOME/.config/transit/config.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "turn on verbose logging")

	// Local to root flags
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "V", false, "print installed version number")

	// Subcommands
	rootCmd.AddCommand(
		a.newAtCmd(),
		a.newConfigCmd(),
		a.newIncidentsCmd(),
		a.newInitCmd(),
	)

	return rootCmd
}

// Run builds the app, runs the command tree, and returns a process exit code.
func Run() int {
	app := &App{Out: os.Stdout, Now: time.Now, Log: logger.New(false)}
	// Too late to change the exit code, but logging to make debugging this scenario
	// easier
	defer func() {
		if err := app.close(); err != nil {
			app.Log.Warn(fmt.Sprintf("Failed to close the database: %s", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.run(ctx, os.Args[1:])
}

// cancelled reports whether err is the user backing out from a prompt or from a signal.
func cancelled(err error) bool {
	return errors.Is(err, ui.ErrCancelled) || errors.Is(err, context.Canceled)
}

// exitCode maps a caught error to one of the documented exit codes.
func exitCode(err error) int {
	var httpErr *provider.HTTPError

	switch {
	case err == nil:
		return 0
	case cancelled(err):
		return 0 // Acceptable errors that aren't real errors
	case errors.Is(err, errUsage),
		errors.Is(err, provider.ErrMissingAPIKey),
		errors.Is(err, ui.ErrNoSelection),
		errors.Is(err, ui.ErrNoInput),
		errors.Is(err, config.ErrInvalid):
		return 2 // Usage or configuration error
	case errors.As(err, &httpErr):
		return 3 // Network or upstream error
	case errors.Is(err, provider.ErrNoDepartures):
		return 4 // Request was successful but there's nothing to show
	default:
		return 1 // Catch-all, internal error
	}
}
