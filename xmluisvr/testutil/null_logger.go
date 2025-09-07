// Package testutil provides testing utilities and helpers for the Scout MCP server.
// This file contains logger implementations and utilities for testing scenarios
// including null handlers and quiet logging for tests that don't require log inspection.
package testutil

import (
	"context"
	"log/slog"
)

// NewNullLogger creates a test logger for use in unit tests.
// Currently returns a quiet logger that discards output, but may be enhanced
// in the future to provide buffering capabilities for log inspection.
func NewNullLogger() *slog.Logger {
	return NullLogger() // TODO: Replace this with a buffering logger
}

// NullLogger creates a logger that discards all output (for tests that don't need log inspection)
func NullLogger() *slog.Logger {
	return slog.New(NullHandler{})
}

// NullHandler implements slog.Handler interface by discarding all log output.
// This handler is used in testing scenarios where log output is not needed.
type NullHandler struct{}

// Enabled implements slog.Handler interface and always returns false to disable all logging.
func (NullHandler) Enabled(context.Context, slog.Level) bool { return false }

// Handle implements slog.Handler interface and discards all log records.
func (NullHandler) Handle(context.Context, slog.Record) error { return nil }

// WithAttrs implements slog.Handler interface and returns a new NullHandler ignoring attributes.
func (NullHandler) WithAttrs([]slog.Attr) slog.Handler { return NullHandler{} }

// WithGroup implements slog.Handler interface and returns a new NullHandler ignoring group names.
func (NullHandler) WithGroup(string) slog.Handler { return NullHandler{} }

//// NullLogger creates a logger that discards all output (for tests that don't need log inspection)
//func NullLogger() *slog.Logger {
//	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
//		Level: slog.LevelError + 1, // Set level higher than any used level to discard everything
//	}))
//}
