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
		logger.Error("No api key found in config at 'dmv.api_key'")
		helpers.Exit(helpers.EXIT_BAD_CONFIG)
	}

	return &DmvApi{
		apiKey:  apiKey,
		baseUrl: DMV_BASE_URL,
	}
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
