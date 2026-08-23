package render_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

var update = flag.Bool("update", false, "update golden files")

// Both settings are read when a style renders, not when it's built, so the package-level styles
// in render pick them up. Without them the output depends on the terminal running the test.
func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	os.Exit(m.Run())
}

func golden(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", name+".golden")

	if *update {
		// The first run of -update has no testdata directory to write into.
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("expected to create %s but got %v", filepath.Dir(path), err)
		}

		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("expected to write %s but got %v", path, err)
		}

		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected to read %s but got %v, run `just golden ./internal/render`", path, err)
	}

	want := string(b)
	if got == want {
		return
	}

	n := firstDiff(want, got)

	t.Errorf("expected %q but got %q\n%s line %d: %q",
		around(want, n), around(got, n),
		path, strings.Count(want[:n], "\n")+1, ansi.Strip(lineAt(want, n)))
}

// firstDiff returns the offset of the first byte the two disagree on.
func firstDiff(want, got string) int {
	n := 0
	for n < len(want) && n < len(got) && want[n] == got[n] {
		n++
	}

	return n
}

// Bytes either side of the first difference. A truecolor background escape is 17 of them.
const window = 24

func around(s string, n int) string {
	return s[max(n-window, 0):min(n+window, len(s))]
}

// lineAt returns the line the offset falls on.
func lineAt(s string, n int) string {
	start := strings.LastIndex(s[:n], "\n") + 1

	end := strings.IndexByte(s[n:], '\n')
	if end < 0 {
		return s[start:]
	}

	return s[start : n+end]
}
