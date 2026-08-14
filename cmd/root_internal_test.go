package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/ui"
	"github.com/ismailshak/transit/pkg/api"
)

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  error
		want int
	}{
		"no error": {
			err:  nil,
			want: 0,
		},
		"cancelled prompt": {
			// cmd/init.go wraps whatever executeInitConfig returns
			err:  fmt.Errorf("collect information: %w", ui.ErrCancelled),
			want: 0,
		},
		"unknown flag": {
			// the shape SetFlagErrorFunc produces
			err:  fmt.Errorf("%w: %w", errUsage, errors.New("unknown flag: --nope")),
			want: 2,
		},
		"wrong argument count": {
			// the shape usageArgs produces
			err:  fmt.Errorf("%w: %w", errUsage, errors.New("requires at least 1 arg(s), only received 0")),
			want: 2,
		},
		"missing api key": {
			err:  fmt.Errorf("dmv: %w", api.ErrMissingAPIKey),
			want: 2,
		},
		"nothing selected": {
			err:  fmt.Errorf("collect information: %w", ui.ErrNoSelection),
			want: 2,
		},
		"no input": {
			err:  fmt.Errorf("collect information: %w", ui.ErrNoInput),
			want: 2,
		},
		"unreadable config": {
			err: fmt.Errorf("load config: %w",
				fmt.Errorf("%w: %w", config.ErrInvalid, errors.New("yaml: line 3: mapping values are not allowed"))),
			want: 2,
		},
		"unsupported location": {
			err:  fmt.Errorf("%w: unsupported location %q", config.ErrInvalid, "abc"),
			want: 2,
		},
		"upstream rejected the key": {
			err: fmt.Errorf("at: %w",
				fmt.Errorf("fetch predictions for 902101: %w",
					&api.HTTPError{StatusCode: http.StatusUnauthorized, URL: "http://api.511.org/transit/StopMonitoring"})),
			want: 3,
		},
		"upstream unavailable": {
			err:  fmt.Errorf("fetch incidents: %w", &api.HTTPError{StatusCode: http.StatusServiceUnavailable}),
			want: 3,
		},
		"no departures anywhere": {
			err:  fmt.Errorf("at: %w", api.ErrNoDepartures),
			want: 4,
		},
		"upstream failure outranks no departures": {
			err:  errors.Join(api.ErrNoDepartures, &api.HTTPError{StatusCode: http.StatusServiceUnavailable}),
			want: 3,
		},
		"unclassified": {
			err:  fmt.Errorf("synchronize migrations: %w", errors.New("corrupt migrations (out of sync)")),
			want: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := exitCode(tc.err)

			if got != tc.want {
				t.Errorf("exitCode(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestExitCodeUnwrapsHTTPError(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("at: %w", &api.HTTPError{StatusCode: http.StatusUnauthorized})

	var httpErr *api.HTTPError
	if !errors.As(wrapped, &httpErr) {
		t.Fatalf("errors.As(%v, **api.HTTPError) = false, want true", wrapped)
	}

	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", httpErr.StatusCode, http.StatusUnauthorized)
	}
}
