package cli

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Set by goreleaser at release time. To mimic it locally:
//
// go build -ldflags "-X github.com/ismailshak/transit/internal/cli.version=1.2.3" -o transit
// TODO: Do we still need -ldflags via goreleaser with these new tools?
var version = "dev"

// buildVersion reports the release this binary came from. A `go install` build
// doesn't use ldflags so the module version the toolchain stored is the only source.
func buildVersion() string {
	if version != "dev" {
		return version
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok || bi.Main.Version == "" || bi.Main.Version == "(devel)" {
		return version
	}

	// goreleaser's version has no leading v. Trim so both paths print the same shape.
	return strings.TrimPrefix(bi.Main.Version, "v")
}

func (a *App) executeVersion() {
	_, _ = fmt.Fprintln(a.Out, "transit "+buildVersion())

	if !a.verbose {
		return
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var revision, committed string
	var dirty bool // any uncommitted change, an untracked file included

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.time":
			committed = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}

	// A binary installed with `go install` was built from the module cache, not a checkout.
	if revision == "" {
		return
	}

	if dirty {
		revision += " (dirty)"
	}

	_, _ = fmt.Fprintln(a.Out, "commit: "+revision)
	_, _ = fmt.Fprintln(a.Out, "committed: "+committed)
}
