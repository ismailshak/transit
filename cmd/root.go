package cmd

import (
	"fmt"
	"os"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/internal/version"
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

		if configFile != "" {
			config.SetCustomConfigPath(configFile)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			version.Execute()
			utils.Exit(utils.EXIT_SUCCESS)
		}

		cmd.Help()
	},
}

func init() {
	// Global, persistent flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (defaults to $HOME/.config/transit/config.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "turn on verbose logging")

	// Local to root flags
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "V", false, "print installed version number")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1) // TODO: exit code
	}
}

func dbSetupPreRun(cmd *cobra.Command, args []string) {
	db, err := data.GetDBConn()
	if err != nil {
		logger.Error("Failed to connect to database: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	err = db.SyncMigrations()
	if err != nil {
		logger.Error("Database sync failed: " + err.Error())
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}

func configSetupPreRun(cmd *cobra.Command, args []string) {
	err := config.LoadConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to load config file: %s", err))
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}

func defaultPreRun(cmd *cobra.Command, args []string) {
	configSetupPreRun(cmd, args)
	dbSetupPreRun(cmd, args)
}
