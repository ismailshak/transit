package cmd

import (
	"os"
	"time"

	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/internal/version"
	"github.com/spf13/cobra"
)

func (a *App) newRootCmd() *cobra.Command {
	var versionFlag bool // TODO; make -v the version flag
	var verboseFlag bool // TODO; turn this into a --level flag or keep as verbose?

	rootCmd := &cobra.Command{
		Use:   "transit",
		Short: "Tool for interacting with local transit information",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			a.Log = logger.New(verboseFlag)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				version.Execute()
				utils.Exit(utils.EXIT_SUCCESS)
			}

			if err := cmd.Help(); err != nil {
				a.Log.Error(err.Error())
			}
		},
	}

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	app := &App{Out: os.Stdout, Now: time.Now}

	err := app.newRootCmd().Execute()
	if err != nil {
		os.Exit(1) // TODO: exit code
	}
}
