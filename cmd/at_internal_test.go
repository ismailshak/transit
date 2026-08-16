package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/pkg/api"
)

func TestWatchInterval(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		seconds  int
		expected time.Duration
		err      error
	}{
		"positive seconds converts to a duration": {
			seconds:  30,
			expected: 30 * time.Second,
		},
		"zero is invalid": {
			seconds: 0,
			err:     config.ErrInvalid,
		},
		"negative is invalid": {
			seconds: -1,
			err:     config.ErrInvalid,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := watchInterval(tc.seconds)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error wrapping %v but got %v", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			if got != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, got)
			}
		})
	}
}

func TestEndsWatch(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		err      error
		expected bool
	}{
		"no departures keeps watching": {
			err:      api.ErrNoDepartures,
			expected: false,
		},
		"429 keeps watching": {
			err:      &api.HTTPError{StatusCode: http.StatusTooManyRequests},
			expected: false,
		},
		"500 keeps watching": {
			err:      &api.HTTPError{StatusCode: http.StatusInternalServerError},
			expected: false,
		},
		"503 keeps watching": {
			err:      &api.HTTPError{StatusCode: http.StatusServiceUnavailable},
			expected: false,
		},
		"wrapped 500 keeps watching": {
			err:      fmt.Errorf("fetch predictions for %q: %w", "courth", &api.HTTPError{StatusCode: http.StatusInternalServerError}),
			expected: false,
		},
		"404 ends the watch": {
			err:      &api.HTTPError{StatusCode: http.StatusNotFound},
			expected: true,
		},
		"401 ends the watch": {
			err:      &api.HTTPError{StatusCode: http.StatusUnauthorized},
			expected: true,
		},
		"a dns failure during a dial keeps watching": {
			err: fetchErr(&net.OpError{
				Op:  "dial",
				Err: &net.DNSError{Err: "no such host", Name: "api.wmata.com", IsNotFound: true},
			}),
			expected: false,
		},
		// The only case that reaches the DNSError arm.
		"a bare resolver failure keeps watching": {
			err:      fetchErr(&net.DNSError{Err: "no such host", Name: "api.wmata.com", IsNotFound: true}),
			expected: false,
		},
		"a refused connection keeps watching": {
			err:      fetchErr(&net.OpError{Op: "dial", Err: errors.New("connect: connection refused")}),
			expected: false,
		},
		// net/http builds its Client.Timeout error from an unexported type, but it
		// reports itself as DeadlineExceeded.
		"a timeout keeps watching": {
			err:      fetchErr(context.DeadlineExceeded),
			expected: false,
		},
		"a certificate failure ends the watch": {
			err: fetchErr(&tls.CertificateVerificationError{
				Err: errors.New("x509: certificate signed by unknown authority"),
			}),
			expected: true,
		},
		"a cancelled request ends the watch": {
			err:      fetchErr(context.Canceled),
			expected: true,
		},
		"a cancel during the dial ends the watch": {
			err:      fetchErr(&net.OpError{Op: "dial", Err: context.Canceled}),
			expected: true,
		},
		// Ctrl-C while the resolver is stalled (which reaches the DNSError arm too).
		"a cancel during the lookup ends the watch": {
			err: fetchErr(&net.DNSError{
				Err:       "operation was canceled",
				Name:      "api.wmata.com",
				UnwrapErr: context.Canceled,
			}),
			expected: true,
		},
		"a missing api key ends the watch": {
			err:      api.ErrMissingAPIKey,
			expected: true,
		},
		"a config error ends the watch": {
			err:      fmt.Errorf("%w: watch_interval must be greater than 0", config.ErrInvalid),
			expected: true,
		},
		"a usage error ends the watch": {
			err:      fmt.Errorf("%w: no station matched %v", errUsage, []string{"nonsense"}),
			expected: true,
		},
		"a malformed payload ends the watch": {
			err:      parseErr(),
			expected: true,
		},
		"any other error ends the watch": {
			err:      errors.New("oops"),
			expected: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := endsWatch(tc.err); got != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, got)
			}
		})
	}
}

func fetchErr(cause error) error {
	urlErr := &url.Error{Op: "Get", URL: "https://api.wmata.com/StationPrediction.svc/json/GetPrediction", Err: cause}
	return fmt.Errorf("fetch predictions for %q: %w", "courth", urlErr)
}

// parseErr is what a truncated response body produces inside FetchPredictions.
func parseErr() error {
	err := json.Unmarshal([]byte("{"), &struct{}{})
	return fmt.Errorf("parse predictions response: %w", err)
}
