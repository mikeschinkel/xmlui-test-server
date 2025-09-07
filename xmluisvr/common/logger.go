package common

import (
	"log/slog"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}

func ensureLogger() *slog.Logger {
	if logger == nil {
		panic("Must call common.SetLogger() with a *slog.Logger before using common package")
	}
	return logger
}
