/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/list"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list <args>",
	Example: "  transit list courthouse (matches \"Court House\")\n  transit list metro (matches \"Metro Center\")",
	Short:   "Display next train arrival information for chosen station(s)",
	Long: `
'list' will display arriving train information for one or more stations.

Arguments are considered valid if it can be used to narrow 
the official station names to just 1. If something's too generic,
try being more specific by adding more characters.
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.DmvClient()
		list.Execute(client, args)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
