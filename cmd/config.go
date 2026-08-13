package cmd

import (
	"fmt"
	"strconv"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/utils"
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
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun:                a.defaultPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			value := a.executeGet(args[0])
			if value == "" {
				a.Log.Warn("No config property found matching '", args[0], "'")
				return
			}

			_, _ = fmt.Fprintln(a.Out, value)
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
		Args:                  cobra.ExactArgs(2),
		PreRun:                a.defaultPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			err := a.executeSet(args[0], args[1])
			if err != nil {
				a.Log.Error(err.Error())
				utils.Exit(utils.EXIT_BAD_CONFIG)
			}

			_, _ = fmt.Fprintf(a.Out, "'%s' has been set to '%s'\n", args[0], args[1])
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
		Args:                  cobra.NoArgs,
		PreRun:                a.defaultPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			a.executePath()
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
func (a *App) executeSet(key, value string) error {
	valid := a.validateKey(key, value)
	if !valid {
		return fmt.Errorf("invalid value for key '%s'", key)
	}

	return a.Cfg.Set(key, value)
}

// executePath backs `config path`
func (a *App) executePath() {
	_, _ = fmt.Fprintln(a.Out, a.Cfg.FileUsed())
}

func (a *App) validateKey(key, value string) bool {
	switch key {
	case "core.location":
		return a.validateLocation(value)
	case "core.watch_interval":
		return a.validateWatchInterval(value)
	}

	return true
}

func (a *App) validateLocation(location string) bool {
	l, err := a.Store.GetLocation(data.LocationSlug(location))
	if err != nil {
		a.Log.Error(fmt.Sprintf("Failed to validate location: %s", location))
		return false
	}

	if l == nil {
		a.Log.Error(fmt.Sprintf("'%s' is not a valid location", location))
		return false
	}

	return true
}

func (a *App) validateWatchInterval(interval string) bool {
	i, err := strconv.ParseInt(interval, 10, 0)
	if err != nil {
		a.Log.Error("'watch_interval' value must be an integer")
		return false
	}

	if i <= 0 {
		a.Log.Error("'watch_interval' value must be greater than 0")
		return false
	}

	return true
}
