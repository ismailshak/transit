package api

import (
	"fmt"

	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
)

const (
	DMV_BASE_URL = "https://api.wmata.com"
)

// Next train arrival prediction data
type Predictions struct {
	// Minutes until a train arrives
	Min string
	// The short name for a train station
	LocationName string
	// The short name for the train's destination
	Destination string
	// The full name for the train's destination
	DestinationName string
	// The train's line
	Line string
}

type Api interface {
	// Fetches arrival information for list of location unique identifiers
	FetchPredictions(ids []string) ([]Predictions, error)
	// Given user input for a location, returns the unique identifier (location could have multiple)
	GetCodeFromArg(arg string) []string
}

func GetClient(location string) Api {
	if location == "" {
		logger.Error("No location found in config at 'core.location'")
		return nil
	}

	switch location {
	case "dmv":
		return DmvClient()
	default:
		logger.Error(fmt.Sprintf("Invalid location '%s'", location))
	}

	return nil
}

// Build and return a client for the DMV Metro
func DmvClient() *DmvApi {
	apiKey := &config.GetConfig().Dmv.ApiKey

	if *apiKey == "" {
		logger.Error("No api key found in config at 'dmv.api_key'")
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	return &DmvApi{
		apiKey:  apiKey,
		baseUrl: DMV_BASE_URL,
	}
}
