package tui

import "github.com/ismailshak/transit/internal/logger"

func OperationSuccessful(msg string) {
	logger.Print(OP_SUCCESS_STYLE(SUCCESS_ICON), msg)
}

func OperationFailed(msg string) {
	logger.Print(OP_FAILED_STYLE(ERROR_ICON), msg)
}

func OperationSkipped(msg string) {
	logger.Print(OP_SKIPPED_STYLE(SKIP_ICON), msg)
}
