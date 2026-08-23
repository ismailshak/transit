package tui_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/ismailshak/transit/internal/tui"
)

// Every Terminal below wraps a bytes.Buffer, which has no file descriptor and so
// is never a terminal.

func TestWithSpinnerWithoutTerminal(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	calls := 0

	err := tui.NewTerminal(&out).WithSpinner(t.Context(), &tui.SpinnerOptions{
		SpinMessage: "Fetching data...",
		CallbackFn: func(context.Context) error {
			calls++
			return nil
		},
	})
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if calls != 1 {
		t.Errorf("expected the callback to run once but it ran %d times", calls)
	}

	if got := out.String(); got != "Fetching data...\n" {
		t.Errorf("expected %q but got %q", "Fetching data...\n", got)
	}
}

func TestWithSpinnerWithoutTerminalReturnsTheCallbackError(t *testing.T) {
	t.Parallel()

	want := errors.New("fetch failed")
	err := tui.NewTerminal(&bytes.Buffer{}).WithSpinner(t.Context(), &tui.SpinnerOptions{
		SpinMessage: "Fetching data...",
		CallbackFn:  func(context.Context) error { return want },
	})

	if !errors.Is(err, want) {
		t.Errorf("expected %v but got %v", want, err)
	}
}

// ctx is the only way out early and it can only arrive as ErrCancelled.
func TestWithSpinnerWithoutTerminalCancels(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := tui.NewTerminal(&bytes.Buffer{}).WithSpinner(ctx, &tui.SpinnerOptions{
		SpinMessage: "Fetching data...",
		CallbackFn:  func(ctx context.Context) error { return ctx.Err() },
	})

	if !errors.Is(err, tui.ErrCancelled) {
		t.Errorf("expected %v but got %v", tui.ErrCancelled, err)
	}
}

// prompts refuse instead of degrading (like the spinner degrades).
func TestPromptsWithoutTerminal(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*testing.T, *tui.Terminal) (string, error){
		"select": func(t *testing.T, s *tui.Terminal) (string, error) {
			return s.Select(t.Context(), "Select a location", []tui.Choice{{Key: "dmv"}})
		},
		"password": func(t *testing.T, s *tui.Terminal) (string, error) {
			return s.Password(t.Context(), "Enter your API key")
		},
	}

	for name, ask := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			got, err := ask(t, tui.NewTerminal(&out))

			if !errors.Is(err, tui.ErrNotInteractive) {
				t.Errorf("expected %v but got %v", tui.ErrNotInteractive, err)
			}

			if got != "" {
				t.Errorf("expected no answer but got %q", got)
			}

			if out.Len() != 0 {
				t.Errorf("expected nothing written but got %q", out.String())
			}
		})
	}
}
