package tui

import (
	"fmt"
)

func OperationSuccessful(msg string) {
	fmt.Println(OP_SUCCESS_STYLE(SUCCESS_ICON), msg)
}

func OperationFailed(msg string) {
	fmt.Println(OP_FAILED_STYLE(ERROR_ICON), msg)
}

func OperationSkipped(msg string) {
	fmt.Println(OP_SKIPPED_STYLE(SKIP_ICON), msg)
}
