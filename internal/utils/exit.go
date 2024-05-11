package utils

import "os"

type ExitCode = int

const (
	EXIT_SUCCESS    ExitCode = 0
	EXIT_FAILURE    ExitCode = 1
	EXIT_BAD_USAGE  ExitCode = 64
	EXIT_BAD_CONFIG ExitCode = 78
)

// Wrapper around [os.Exit]
func Exit(code ExitCode) {
	os.Exit(code)
}
