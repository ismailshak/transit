/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
	"github.com/ismailshak/transit/version"
	"github.com/spf13/cobra"
)

// Used for flags
var (
	configFile  string
	versionFlag bool
	verboseFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "transit",
	Short: "Tool for interacting with local transit information",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verboseFlag {
			configFile := config.GetConfig()
			configFile.Core.Verbose = true
		}
	},
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
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "toggle verbose logging")

	// Local to root flags
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "print installed version number")
}

func initConfig() {
	LoadConfig(configFile)
}

// This should be called before any command gets parsed & executed
func LoadConfig(path string) {
	configFile := config.GetConfig()
	if path != "" {
		logger.Debug(fmt.Sprintf("Config path override provided: '%s'", path))
		vp.SetConfigFile(path)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error(fmt.Sprint(err))
			helpers.Exit(helpers.EXIT_BAD_CONFIG)
		}

		err = helpers.CreatePathIfNotFound(homeDir + "/.config/transit/config.yml")
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create config directory: %s", err))
			helpers.Exit(helpers.EXIT_BAD_CONFIG)
		}

		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(homeDir + "/.config/transit/")
	}

	err := vp.ReadInConfig()
	if err != nil {
		logger.Error(fmt.Sprint(err))
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	err = vp.Unmarshal(&configFile)
	if err != nil {
		logger.Error("Failed to parse config" + fmt.Sprint(err))
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	// TODO: we probably don't need to do this, given its already a pointer
	config.SetConfig(configFile)
}
