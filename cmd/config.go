package cmd

import (
	"fmt"

	"github.com/ismailshak/transit/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config <command>",
	Short: "Manage configuration for transit CLI",
	Long: `
Get and set configuration options.

For nested config options, use a period/full-stop/dot as a delimiter.

There is very minimal validation on the values you set. Whatever you decide
to set will probably be written... will just not be parsed/read by any of the commands.`,
	DisableFlagsInUseLine: true,
}

var configGetCommand = &cobra.Command{
	Use:                   "get <key>",
	Short:                 "Get a key from the config file",
	Long:                  "Get a key from the config file\nFor all values, check the README at https://github.com/ismailshak/transit#config",
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
	Long:                  "Set a key in the config file\nFor all values, check the README at https://github.com/ismailshak/transit",
	Example:               "  transit config set core.location dmv\n  transit config set dmv.api_key abcdef",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSet(args[0], args[1])
	},
}

var configPathCommand = &cobra.Command{
	Use:                   "path",
	Short:                 "Prints path to configuration file",
	Long:                  "Prints path to configuration file",
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

var (
	vp *viper.Viper = viper.New()
)

// Entry point for `config get`
func ExecuteGet(key string) {
	result := vp.Get(key)

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

	vp.Set(key, value)
	vp.WriteConfig()
	logger.Print(fmt.Sprintf("'%s' has been set to '%s'\n", key, value))
}

// Entry point for `config path`
func ExecutePath() {
	logger.Print(getConfigPath())
}

func validateKey(key, value string) bool {
	if key == "core.location" {
		valid := validateLocation(value)
		if !valid {
			logger.Error(fmt.Sprintf("'%s' is not a valid location\n", value))
			return false
		}
	}

	return true
}

var validLocations = map[string]bool{
	"dmv": true,
}

func validateLocation(location string) bool {
	_, exists := validLocations[location]
	return exists
}

func getConfigPath() string {
	return vp.ConfigFileUsed()
}
