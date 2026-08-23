package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ismailshak/transit/internal/transit"
	"github.com/ismailshak/transit/internal/tui"
)

func TestToChoices(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    []transit.Location
		expected []tui.Choice
	}{
		"converts a location correctly": {
			input: []transit.Location{
				{Slug: "slug", Name: "Long Slug Name"},
			},
			expected: []tui.Choice{
				{Key: "slug", Title: "slug", Description: "Long Slug Name", FilterValue: "Long Slug Name"},
			},
		},
		"empty list": {
			input:    []transit.Location{},
			expected: []tui.Choice{},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := toChoices(testCase.input)

			if !slices.Equal(got, testCase.expected) {
				t.Errorf("expected %v but got %v", testCase.expected, got)
			}
		})
	}
}

// newTestApp writes to a bytes.Buffer, so these run the same path as a piped `transit init`.
func TestInitWithoutTerminal(t *testing.T) {
	t.Run("location picker mentions resolution", func(t *testing.T) {
		app := newTestApp(t)

		code := app.run("init")
		if code != 2 {
			t.Fatalf("expected exit code 2 but got %d (output %q)", code, app.out)
		}

		if !strings.Contains(app.err.String(), "transit config set core.location") {
			t.Errorf("expected the alternative on Err but got %q", app.err)
		}

		if strings.Contains(app.out.String(), "\x1b") {
			t.Errorf("expected no escape sequences on Out but got %q", app.out)
		}
	})

	t.Run("api key prompt mentions resolution", func(t *testing.T) {
		app := newTestApp(t)
		seedLocation(t, app.home, "dmv")

		code := app.run("init")
		if code != 2 {
			t.Fatalf("expected exit code 2 but got %d (output %q)", code, app.out)
		}

		if !strings.Contains(app.err.String(), "transit config set dmv.api_key") {
			t.Errorf("expected the alternative on Err but got %q", app.err)
		}
	})
}

// seedLocation writes what `config set` would write. That command opens the store a second time.
// Windows then won't let TempDir remove the handle.
// TODO: go through the CLI once the store and the config have their own flag-exposed paths
func seedLocation(t *testing.T, home, location string) {
	t.Helper()

	dir := filepath.Join(home, ".config", "transit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	contents := fmt.Sprintf("core:\n  location: %s\n", location)
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(contents), 0o600); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
}
