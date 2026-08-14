package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// errUsage marks an invalid invocation: an unknown flag, or the wrong number of
// arguments. Callers wrap it with the specifics. Mapped to exit code 2
var errUsage = errors.New("usage")

// usageArgs tags a cobra argument validator's failures with errUsage, so a bad
// invocation reaches exitCode as a value rather than a bare string. Cobra builds
// these errors itself, which is why they have to be tagged on the way out rather
// than raised where the condition is detected.
func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return fmt.Errorf("%w: %w", errUsage, err)
		}

		return nil
	}
}
