// Package provider contains clients that manage transit data hosted by transit agencies.
//
// Each supported location implements the interface [transit.Provider], which encapsulates the specifics of fetching
// and parsing data coming from a specific transit data provider.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ismailshak/transit/internal/transit"
)

const (
	wmataBaseURL  = "https://api.wmata.com"
	sfBaseURL     = "http://api.511.org"
	httpTimeout   = 15 * time.Second
	wmataTimezone = "America/New_York"
)

// staticLookup is the parts of the store the clients read.
// Makes testing easier.
type staticLookup interface {
	Agencies(ctx context.Context, location transit.LocationSlug) ([]transit.Agency, error)
}

// NewDMV builds a client for the DMV Metro Area, backed by WMATA.
func NewDMV(apiKey string, now func() time.Time) (*WMATAClient, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	location, err := time.LoadLocation(wmataTimezone)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", wmataTimezone, err)
	}

	return &WMATAClient{
		apiKey:   apiKey,
		baseURL:  wmataBaseURL,
		http:     &http.Client{Timeout: httpTimeout},
		location: location,
		now:      now,
	}, nil
}

// NewSF builds a client for the San Francisco Bay Area, backed by 511.
func NewSF(apiKey string, s staticLookup) (*SFClient, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	return &SFClient{
		apiKey:  apiKey,
		baseURL: sfBaseURL,
		http:    &http.Client{Timeout: httpTimeout},
		store:   s,
	}, nil
}
