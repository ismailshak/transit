package ui

import (
	"bytes"
	"testing"
)

// terminalScreen returns a Screen that forces its buffer to be a terminal.
func terminalScreen() (*Terminal, *bytes.Buffer) {
	var out bytes.Buffer
	return &Terminal{w: &out, fd: -1, tty: true}, &out
}

func TestDrawOnTerminal(t *testing.T) {
	t.Parallel()

	s, out := terminalScreen()
	s.Draw("frame")

	want := cursorHome + "frame" + eraseToEnd
	if got := out.String(); got != want {
		t.Errorf("expected %q but got %q", want, got)
	}
}

func TestEnterOnTerminal(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		enters int
		want   string
	}{
		"once": {
			enters: 1,
			want:   openAltBuffer,
		},
		"twice opens one buffer": {
			enters: 2,
			want:   openAltBuffer,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s, out := terminalScreen()
			for range tt.enters {
				if err := s.Enter(); err != nil {
					t.Fatalf("expected no error but got %v", err)
				}
			}

			if got := out.String(); got != tt.want {
				t.Errorf("expected %q but got %q", tt.want, got)
			}
		})
	}
}

func TestExitOnTerminal(t *testing.T) {
	t.Parallel()

	t.Run("closes a buffer that was never opened", func(t *testing.T) {
		t.Parallel()

		s, out := terminalScreen()
		if err := s.Exit(); err != nil {
			t.Fatalf("expected no error but got %v", err)
		}

		if got := out.String(); got != closeAltBuffer {
			t.Errorf("expected %q but got %q", closeAltBuffer, got)
		}
	})

	t.Run("leaves the screen able to enter again", func(t *testing.T) {
		t.Parallel()

		s, out := terminalScreen()
		for _, step := range []func() error{s.Enter, s.Exit, s.Enter} {
			if err := step(); err != nil {
				t.Fatalf("expected no error but got %v", err)
			}
		}

		want := openAltBuffer + closeAltBuffer + openAltBuffer
		if got := out.String(); got != want {
			t.Errorf("expected %q but got %q", want, got)
		}
	})
}
