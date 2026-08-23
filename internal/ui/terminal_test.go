package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ismailshak/transit/internal/ui"
)

// A bytes.Buffer has no file descriptor, so every Terminal below is a non-terminal.
func TestScreenWithoutTerminal(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		do   func(*ui.Terminal)
		want string
	}{
		"enter": {
			do:   func(s *ui.Terminal) { _ = s.Enter() },
			want: "",
		},
		"exit": {
			do:   func(s *ui.Terminal) { _ = s.Exit() },
			want: "",
		},
		"draw": {
			do:   func(s *ui.Terminal) { s.Draw("one\ntwo") },
			want: "one\ntwo",
		},
		"print": {
			do:   func(s *ui.Terminal) { s.Print("hello") },
			want: "hello",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			tc.do(ui.NewTerminal(&out))

			if got := out.String(); got != tc.want {
				t.Errorf("expected %q but got %q", tc.want, got)
			}
		})
	}
}

func TestScreenWithoutTerminalEmitsNoEscapes(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	s := ui.NewTerminal(&out)

	if err := s.Enter(); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	s.Draw("first frame")
	s.Draw("second frame")
	s.Print("done")

	if err := s.Exit(); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if got := out.String(); strings.Contains(got, "\x1b") {
		t.Errorf("expected no escape sequences but got %q", got)
	}
}

func TestScreenWithoutTerminalHasNoWidth(t *testing.T) {
	t.Parallel()

	s := ui.NewTerminal(&bytes.Buffer{})

	if s.TTY() {
		t.Error("expected no terminal but got one")
	}

	if got := s.Width(); got != 0 {
		t.Errorf("expected 0 but got %d", got)
	}
}
