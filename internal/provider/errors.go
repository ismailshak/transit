package provider

import (
	"errors"
	"fmt"
)

// ErrMissingAPIKey is returned when the configured location has no credentials.
var ErrMissingAPIKey = errors.New("missing api key")

// HTTPError is returned when a transit API responds with an unexpected status code.
type HTTPError struct {
	StatusCode int
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: unexpected status %d", e.URL, e.StatusCode)
}
