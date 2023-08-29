// Package api contains functions that manage transit information exposed by transit agencies.
//
// Each supported location implements the interface `Api` and encapsulates the details of fetching
// and parsing data coming from a specific transit agency. There will exist a client for
// each supported location, that can be retrieved dynamically via `GetClient("locationSlug")`.
package api

import (
	"fmt"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
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
	// Message from the transit authority describing the issue
	Description string
	// When the announcement was last updated by the transit authority
	DateUpdated time.Time
	// Lines or stops affected by the incident
	Affected []string
	// Type of incident (e.g. "alert")
	Type string
}

// Base interface that defines what each location client api must implement
type Api interface {
	// Fetches all required static data. Used to hydrate database
	FetchStaticData() (*data.Data, error)
	// Fetches arrival information for list of location unique identifiers
	FetchPredictions(ids []string) ([]Prediction, error)
	// Fetch all incidents reported by the agency for a location
	FetchIncidents() ([]Incident, error)
	// Given user input for a location, returns the unique identifier (a stop can have multiple)
	GetIDFromArg(arg string) ([]string, error)
	// Given a line name or abbreviation, return colors that represents it.
	// (bg, fg) tuple returned
	GetLineColor(stop string) (string, string)
	// Determines if a train isn't for passengers
	IsGhostTrain(line, destination string) bool
}

// Dynamically retrieve the client associated with the provided location
func GetClient(location data.LocationSlug) Api {
	if location == "" {
		logger.Error("No location found in config at 'core.location'")
		return nil
	}

	switch location {
	case data.DMVSlug:
		return DmvClient()
	default:
		logger.Error(fmt.Sprintf("Invalid location '%s'", location))
	}

	return nil
}

// Build and return a client for the DMV Metro Area
func DmvClient() *DmvApi {
	apiKey := &config.GetConfig().Dmv.ApiKey

	if *apiKey == "" {
		logger.Error("No api key found in config at 'dmv.api_key'")
		utils.Exit(utils.EXIT_BAD_CONFIG)
	}

	return &DmvApi{
		apiKey:  apiKey,
		baseUrl: DMV_BASE_URL,
	}
}
