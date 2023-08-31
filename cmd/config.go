package cmd

import (
	"fmt"
	"strconv"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <command>",
	Short: "Manage configuration for transit CLI",
	Long: `
Get and set configuration options.

For nested config options, use a dot (.) as a delimiter.`,
	DisableFlagsInUseLine: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		defaultPreRun(cmd, args)
	},
}

var configGetCommand = &cobra.Command{
	Use:                   "get <key>",
	Short:                 "Get a key from the config file",
	Long:                  "Get a key from the configuration file\nFor all values, check the docs https://transitcli.com/docs/config-reference",
	Example:               "  transit config get core.location",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteGet(args[0])
	},
}

var configSetCommand = &cobra.Command{
	Use:                   "set <key> <value>",
	Short:                 "Set a key in the config file",
	Long:                  "Set a key in the configuration file\nFor all values, check the docs https://transitcli.com/docs/config-reference",
	Example:               "  transit config set core.location dmv\n  transit config set dmv.api_key abcdef",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSet(args[0], args[1])
	},
}

var configPathCommand = &cobra.Command{
	Use:                   "path",
	Short:                 "Prints path to config file",
	Long:                  "Prints path to configuration file used",
	DisableFlagsInUseLine: true,
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ExecutePath()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCommand)
	configCmd.AddCommand(configSetCommand)
	configCmd.AddCommand(configPathCommand)
}

// Entry point for `config get`
func ExecuteGet(key string) {
	result := config.GetValue(key)

	if result == nil {
		logger.Warn(fmt.Sprintf("No config property found matching '%s'\n", key))
		return
	}

	logger.Print(fmt.Sprint(result))
}

// Entry point for `config set`
func ExecuteSet(key, value string) {
	valid := validateKey(key, value)
	if !valid {
		return
	}

	config.SetValue(key, value)
	logger.Print(fmt.Sprintf("'%s' has been set to '%s'\n", key, value))
}

// Entry point for `config path`
func ExecutePath() {
	logger.Print(config.GetConfigFileUsed())
}

func validateKey(key, value string) bool {
	switch key {
	case "core.location":
		return validateLocation(value)
	case "core.watch_interval":
		return validateWatchInterval(value)
	case "core.verbose":
		return validateVerbose(value)
	}

	return true
}

func validateLocation(location string) bool {
	db, _ := data.GetDBConn()
	l, err := db.GetLocation(data.LocationSlug(location))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to validate location: %s", location))
		return false
	}

	if l != nil {
		logger.Error(fmt.Sprintf("'%s' is not a valid location", location))
		return false
	}

	return true
}

func validateWatchInterval(interval string) bool {
	i, err := strconv.ParseInt(interval, 10, 0)
	if err != nil {
		logger.Error("'watch_interval' value must be an integer")
		return false
	}

	if i <= 0 {
		logger.Error("'watch_interval' value must be greater than 0")
		return false
	}

	return true
}

func validateVerbose(value string) bool {
	if value != "true" && value != "false" {
		logger.Error("'verbose' must be set to either 'true' or 'false'")
		return false
	}

	return true
}
