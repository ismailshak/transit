// Package provider contains functions that manage transit information exposed by transit agencies.
//
// Each supported location implements the interface `API` and encapsulates the details of fetching
// and parsing data coming from a specific transit agency.
package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
)

const (
	wmataBaseURL = "https://api.wmata.com"
	sfBaseURL    = "http://api.511.org"
	httpTimeout  = 15 * time.Second
)

// PredictionInput is data required to make a prediction request
type PredictionInput struct {
	StopID   string
	AgencyID string
}

// Prediction is next train arrival prediction data
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

// Incident is disruptions and/or delays data
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

// API is the base interface that defines what each location client api must implement
type API interface {
	// Fetches all required static data. Used to hydrate database
	FetchStaticData(ctx context.Context) (*transit.StaticData, error)
	// Fetches arrival information for list of location unique identifiers
	FetchPredictions(ctx context.Context, input []PredictionInput) ([]Prediction, error)
	// Fetch all incidents reported by the agency for a location
	FetchIncidents(ctx context.Context) ([]Incident, error)
	// Given user input for a location, returns the formatted input required to make a prediction request
	GetPredictionInput(ctx context.Context, arg string) ([]PredictionInput, error)
	// Given a line name or abbreviation, return colors that represents it.
	// (bg, fg) tuple returned
	GetLineColor(stop string) (string, string)
	// Determines if a train isn't for passengers
	IsGhostTrain(line, destination string) bool
}

// NewDMV builds a client for the DMV Metro Area, backed by WMATA
func NewDMV(apiKey string, s *store.Store) (*WMATAClient, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	return &WMATAClient{
		apiKey:  apiKey,
		baseURL: wmataBaseURL,
		http:    &http.Client{Timeout: httpTimeout},
		store:   s,
	}, nil
}

// NewSF builds a client for the San Francisco Bay Area, backed by 511.
// The now function supplies the clock that arrival times are measured against.
func NewSF(apiKey string, s *store.Store, now func() time.Time) (*SFClient, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	return &SFClient{
		apiKey:  apiKey,
		baseURL: sfBaseURL,
		http:    &http.Client{Timeout: httpTimeout},
		now:     now,
		store:   s,
	}, nil
}
