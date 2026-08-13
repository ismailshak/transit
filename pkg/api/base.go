// Package api contains functions that manage transit information exposed by transit agencies.
//
// Each supported location implements the interface `Api` and encapsulates the details of fetching
// and parsing data coming from a specific transit agency.
package api

import (
	"time"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
)

const (
	DMV_BASE_URL = "https://api.wmata.com"
	SF_BASE_URL  = "http://api.511.org"
)

// Data required to make a prediction request
type PredictionInput struct {
	StopID   string
	AgencyID string
}

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
	// The start date/time of the active period for the incident
	ActivePeriodStart time.Time
	// The end date/time of the active period for the incident
	ActivePeriodEnd time.Time
	// Lines or stops affected by the incident
	Affected []string
	// The name of the transit agency that reported the incident
	Agency string
	// Message from the transit authority describing the issue
	Description string
	// When the announcement was last updated by the transit authority
	DateUpdated time.Time
	// Type of incident (e.g. "alert")
	Type string
}

// Base interface that defines what each location client api must implement
type Api interface {
	// Fetches all required static data. Used to hydrate database
	FetchStaticData() (*data.StaticData, error)
	// Fetches arrival information for list of location unique identifiers
	FetchPredictions(input []PredictionInput) ([]Prediction, error)
	// Fetch all incidents reported by the agency for a location
	FetchIncidents() ([]Incident, error)
	// Given user input for a location, returns the formatted input required to make a prediction request
	GetPredictionInput(arg string) ([]PredictionInput, error)
	// Given a line name or abbreviation, return colors that represents it.
	// (bg, fg) tuple returned
	GetLineColor(stop string) (string, string)
	// Determines if a train isn't for passengers
	IsGhostTrain(line, destination string) bool
}

// NewDMV builds a client for the DMV Metro Area, backed by WMATA
func NewDMV(apiKey string, store *data.TransitDB, log *logger.Logger) *DmvApi {
	return &DmvApi{
		apiKey:  apiKey,
		baseUrl: DMV_BASE_URL,
		log:     log,
		store:   store,
	}
}

// NewSF builds a client for the San Francisco Bay Area, backed by 511.
// The now function supplies the clock that arrival times are measured against.
func NewSF(apiKey string, store *data.TransitDB, log *logger.Logger, now func() time.Time) *SFApi {
	return &SFApi{
		apiKey:  apiKey,
		baseUrl: SF_BASE_URL,
		log:     log,
		now:     now,
		store:   store,
	}
}
