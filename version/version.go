package version

import "github.com/ismailshak/transit/logger"

// These variables are set by build flags (goreleaser injects values at release time)
// To mimic locally:
//   - replace all <path> occurances with github.com/ismailshak/transit/version
//   - replace {{.Var}} variables with hardcoded values
//
// go build -ldflags "-X <path>.version={{.Version}} -X <path>.commit={{.Commit}} -X <path>.date={{.Date}}" -o transit
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Execute() {
	logger.Print(formatVersion())
}

func formatVersion() string {
	return version
}
