package utils

func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}

	return b
}
