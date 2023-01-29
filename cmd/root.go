/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/ismailshak/transit/config"
	"github.com/spf13/cobra"
)

// Used for flags
var (
	configFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "transit",
	Short: "Tool for interacting with local transit information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello")
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
}

func initConfig() {
	config.LoadConfig(&configFile)
	// TODO: handle errors
	// TODO: set some defaults for missing fields
}
