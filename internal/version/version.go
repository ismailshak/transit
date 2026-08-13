// Tiny package that just encapsulates injecting data for the version root flag
//
// These variables in here are set by build flags (goreleaser injects values at release time)
// To mimic locally:
//   - replace all <path> occurrences with github.com/ismailshak/transit/internal/version
//   - replace {{.Var}} variables with hardcoded values that represent "Var"
//
// go build -ldflags "-X <path>.version={{.Version}} -X <path>.commit={{.Commit}} -X <path>.date={{.Date}}" -o transit
package version

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Execute() {
	fmt.Println(formatVersion())
}

func formatVersion() string {
	return version
}
