package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/spf13/cobra"
)

func (a *App) newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage configuration for transit CLI",
		Long: `
Get and set configuration options.

For nested config options, use a dot (.) as a delimiter.`,
		DisableFlagsInUseLine: true,
	}

	// Subcommands
	configCmd.AddCommand(
		a.newConfigGetCmd(),
		a.newConfigSetCmd(),
		a.newConfigPathCmd(),
	)

	return configCmd
}

func (a *App) newConfigGetCmd() *cobra.Command {
	configGetCmd := &cobra.Command{
		Use:                   "get <key>",
		Short:                 "Get a key from the config file",
		Long:                  "Get a key from the configuration file\nFor all values, check the docs https://transitcli.com/docs/config-reference",
		Example:               "  transit config get core.location",
		Args:                  usageArgs(cobra.ExactArgs(1)),
		DisableFlagsInUseLine: true,
		PreRunE:               a.configSetupPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			value := a.executeGet(args[0])
			if value == "" {
				return fmt.Errorf("%w: no value set for %q", config.ErrInvalid, args[0])
			}

			_, err := fmt.Fprintln(a.Out, value)
			return err
		},
	}

	return configGetCmd
}

func (a *App) newConfigSetCmd() *cobra.Command {
	configSetCmd := &cobra.Command{
		Use:                   "set <key> <value>",
		Short:                 "Set a key in the config file",
		Long:                  "Set a key in the configuration file\nFor all values, check the docs https://transitcli.com/docs/config-reference",
		Example:               "  transit config set core.location dmv\n  transit config set dmv.api_key abcdef",
		DisableFlagsInUseLine: true,
		Args:                  usageArgs(cobra.ExactArgs(2)),
		PreRunE:               a.defaultPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := a.executeSet(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(a.Out, "'%s' has been set to '%s'\n", args[0], args[1])
			return err
		},
	}

	return configSetCmd
}

func (a *App) newConfigPathCmd() *cobra.Command {
	configPathCmd := &cobra.Command{
		Use:                   "path",
		Short:                 "Prints path to config file",
		Long:                  "Prints path to configuration file used",
		DisableFlagsInUseLine: true,
		Args:                  usageArgs(cobra.NoArgs),
		PreRunE:               a.configSetupPreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.executePath()
		},
	}

	return configPathCmd
}

// executeGet backs `config get`. Returns an empty string if the key isn't set.
func (a *App) executeGet(key string) string {
	result := a.Cfg.Get(key)

	if result == nil {
		return ""
	}

	return fmt.Sprint(result)
}

// executeSet backs `config set`. Validates the value before writing it.
func (a *App) executeSet(ctx context.Context, key, value string) error {
	err := a.validateKey(ctx, key, value)
	if err != nil {
		return err
	}

	return a.Cfg.Set(key, value)
}

// executePath backs `config path`
func (a *App) executePath() error {
	_, err := fmt.Fprintln(a.Out, a.Cfg.FileUsed())
	return err
}

func (a *App) validateKey(ctx context.Context, key, value string) error {
	switch key {
	case "core.location":
		return a.validateLocation(ctx, value)
	case "core.watch_interval":
		return a.validateWatchInterval(value)
	}

	return nil
}

func (a *App) validateLocation(ctx context.Context, location string) error {
	l, err := a.Store.Location(ctx, transit.LocationSlug(location))
	if err != nil {
		return fmt.Errorf("get location data: %w", err)
	}

	if l == nil {
		return fmt.Errorf("%w: %q is not a valid location", config.ErrInvalid, location)
	}

	return nil
}

func (a *App) validateWatchInterval(interval string) error {
	i, err := strconv.ParseInt(interval, 10, 0)
	if err != nil {
		return fmt.Errorf("%w: watch_interval must be an integer", config.ErrInvalid)
	}

	if i <= 0 {
		return fmt.Errorf("%w: watch_interval must be greater than 0", config.ErrInvalid)
	}

	return nil
}
