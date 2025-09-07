package apipkg

import (
	"log/slog"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}
func ensureLogger() {
	if logger == nil {
		panic("Must call apipkg.SetLogger() with a *slog.Logger before using apipkg package")
	}
}
