package common

import (
	"log/slog"
)

var logger *slog.Logger

func Logger() *slog.Logger {
	return logger
}

func SetLogger(l *slog.Logger) {
	logger = l
}

func EnsureLogger() *slog.Logger {
	if logger == nil {
		panic("Must call common.SetLogger() with a *slog.Logger before reaching this check.")
	}
	return logger
}
