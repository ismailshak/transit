package tui

import (
	"fmt"
)

func OperationSuccessful(msg string) {
	fmt.Println(OpSuccessStyle(SuccessIcon), msg)
}

func OperationFailed(msg string) {
	fmt.Println(OpFailedStyle(ErrorIcon), msg)
}

func OperationSkipped(msg string) {
	fmt.Println(OpSkippedStyle(SkipIcon), msg)
}
