package list

import (
	"fmt"
	"strings"

	"github.com/ismailshak/transit/api"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/tui"
	"github.com/sahilm/fuzzy"
)

// Entry point to the `list` subcommand
func Execute(client api.Api, args []string) {
	keys := helpers.GetDmvStationNames()

	for _, arg := range args {
		handleArg(client, keys, arg)
	}
}

// If a single match was found, print the station's arriving trains
func handleArg(client api.Api, allStations []string, arg string) {
	matches := helpers.FuzzyFind(arg, allStations)
	if matches.Len() == 0 {
		printNotFound(arg)
		return
	}

	if matches.Len() > 1 {
		printTooManyMatches(matches, arg)
		return
	}

	matchedStation := matches[0].Str
	codes, _ := helpers.GetStationCodeFromName(matchedStation)

	data := fetchTrains(client, codes)
	tui.PrintArrivingScreen(data)
}

func fetchTrains(client api.Api, stationCodes []string) []api.Timing {
	timings, err := client.ListTimings(stationCodes)
	if err != nil {
		panic(err) // TODO: error handling
	}

	return timings
}

func printNotFound(arg string) {
	fmt.Printf("- Skipping '%s': could not find a matching station\n", arg)
}

func printTooManyMatches(matches fuzzy.Matches, arg string) {
	var names []string
	for _, m := range matches {
		names = append(names, m.Str)
	}

	fmt.Printf("- Skipping '%s': too many matches found\n", arg)
	fmt.Printf("matches: %s\n", strings.Join(names, " - "))
}
