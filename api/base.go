package api

import (
	"fmt"

	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
)

const (
	DMV_BASE_URL = "https://api.wmata.com"
)

type Timing struct {
	Min             string
	LocationName    string
	Destination     string
	DestinationName string
	Line            string
}

type Api interface {
	ListTimings(station []string) ([]Timing, error)
}

// Build and return a client for the DMV Metro
func DmvClient() *DmvApi {
	apiKey := &config.GetConfig().Dmv.ApiKey

	if *apiKey == "" {
		fmt.Println("No api key defined in config at 'dmv.api_key'")
		fmt.Println("Run 'transit config set dmv.api_key <your_key>'")

		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	return &DmvApi{
		apiKey:  apiKey,
		baseUrl: DMV_BASE_URL,
	}
}
