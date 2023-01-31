/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ismailshak/transit/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <args>",
	Short: "Manage configuration for transit CLI",
	Long: `
Get and set configuration options. 

For nested config options, use a period/dot as a delimiter.

There is very minimal validation on the values you set, whatever you decide
to set will probably be approved... just not parsed or read`,
	DisableFlagsInUseLine: true,
}

var configGetCommand = &cobra.Command{
	Use:                   "get <key>",
	Short:                 "Get a key from the config file",
	Long:                  "Get a key from the config file\nFor all values, check the README at https://github.com/ismailshak/transit",
	Example:               "  transfer config get core.location",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.ExecuteGet(args[0])
	},
}

var configSetCommand = &cobra.Command{
	Use:                   "set <key>=<value>",
	Short:                 "Set a key from the config file",
	Long:                  "Set a key from the config file\nFor all values, check the README at https://github.com/ismailshak/transit",
	Example:               "  transfer config set core.location=dmv\n  transfer config set dmv.api_key=abcdef",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.ExecuteSet(args[0])
	},
}

var configPathCommand = &cobra.Command{
	Use:                   "path",
	Short:                 "Prints path to configuration file",
	Long:                  "Prints path to configuration file",
	DisableFlagsInUseLine: true,
	Args:                  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config.ExecutePath()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCommand)
	configCmd.AddCommand(configSetCommand)
	configCmd.AddCommand(configPathCommand)
}
