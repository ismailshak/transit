/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"sort"

	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/config"
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
	station := "K01" // Hardcoding "Courthouse" for now
	timings, err := client.ListTimings(&station)
	if err != nil {
		panic(err)
	}

	sort.Slice(timings, func(i, j int) bool {
		return timings[i].Destination < timings[j].Destination
	})

	for _, t := range timings {
		fmt.Printf("(%s) %s %smin(s)\n", t.Line, t.Destination, t.Min)
	}
}
