package helpers

import "os"

type ExitCode = int

const (
	EXIT_SUCCESS    ExitCode = 0
	EXIT_BAD_USAGE           = 64
	EXIT_BAD_CONFIG          = 78
)

func Exit(code ExitCode) {
	os.Exit(code)
}
