package api

import (
	"fmt"
	"time"

	"github.com/ismailshak/transit/config"
	"github.com/ismailshak/transit/data"
	"github.com/ismailshak/transit/helpers"
	"github.com/ismailshak/transit/logger"
)

const (
	DMV_BASE_URL = "https://api.wmata.com"
)

// Next train arrival prediction data
type Prediction struct {
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

// Disruptions and/or delays data
type Incident struct {
	// Message from the transit authority describing the situation
	Description string
	// Last date & time update from the transit authority
	DateUpdated time.Time
	// Lines, stations or stops affected by the incident
	Affected []string
	// Type of incident
	Type string
}

type Api interface {
	// Fetches all required static data. Used to hydrate database
	FetchStaticData() (*data.Data, error)
	// Fetches arrival information for list of location unique identifiers
	FetchPredictions(ids []string) ([]Prediction, error)
	// Fetch all incidents reported by the agency for a location
	FetchIncidents() ([]Incident, error)
	// Given user input for a location, returns the unique identifier (a stop can have multiple)
	GetIDFromArg(arg string) ([]string, error)
	// Given a stop name or abbreviation, return colors that represents it.
	// (bg, fg) tuple returned
	GetStopColor(stop string) (string, string)
	// Determines if a train isn't for passengers
	IsGhostTrain(line, destination string) bool
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
