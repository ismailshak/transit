/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display next train arrival information for a given station",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.DmvClient(&config.GetConfig().Core.ApiKey)
		executeList(client, args)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func executeList(client api.Api, args []string) {
	station := "C01,A01" // Hardcoding "Metro Center" for now
	timings, err := client.ListTimings(&station)
	if err != nil {
		panic(err) // TODO: error handling
	}

	tui.PrintArrivingScreen(timings)
}
