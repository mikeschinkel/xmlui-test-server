package dbpkg

import (
	"log/slog"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}

func ensureLogger() {
	if logger == nil {
		panic("Must call dbutil.SetLogger() with a *slog.Logger before using dbutil package")
	}
}
