package config

import "errors"

// ErrInvalid marks a configuration problem the user has to fix (an unknown key,
// an unusable value, or a location transit doesn't support). Callers wrap it with
// the specifics. Mapped to exit code 2
var ErrInvalid = errors.New("config")
