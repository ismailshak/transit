package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ismailshak/transit/internal/config"
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			if bi, ok := debug.ReadBuildInfo(); ok {
				fmt.Printf("%+v\n", bi)
			}
			version.Execute()
			utils.Exit(utils.EXIT_SUCCESS)
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
		vp.SetConfigFile(path)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error(fmt.Sprint(err))
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		fullConfigPath := homeDir + "/.config/transit/config.yml"
		err = utils.CreatePathIfNotFound(fullConfigPath)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create config directory: %s", err))
			utils.Exit(utils.EXIT_BAD_CONFIG)
		}

		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(homeDir + "/.config/transit/")
	}

	// config defaults
	vp.SetDefault("core.watch_interval", 10)

	err := vp.ReadInConfig()
	if err != nil {
		logger.Error(fmt.Sprint(err))
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	err = vp.Unmarshal(&configFile)
	if err != nil {
		logger.Error("Failed to parse config\n" + fmt.Sprint(err))
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}
}
