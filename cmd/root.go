/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/version"
	"github.com/spf13/cobra"
)

// Used for flags
var (
	configFile  string
	versionFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "transit",
	Short: "Tool for interacting with local transit information",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			version.Execute()
			helpers.Exit(helpers.EXIT_SUCCESS)
		}

		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global, persistent flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.config/transit/config.yml)")

	// Local to root flags
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "print installed version number")
}

func initConfig() {
	config.LoadConfig(&configFile)
	// TODO: handle errors
	// TODO: set some defaults for missing fields
}
