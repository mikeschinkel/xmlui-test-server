// Package cliutil provides output management and synchronized writing for CLI applications.
package cliutil

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Writer defines the interface for user-facing writer
type Writer interface {
	Printf(string, ...any)
	Errorf(string, ...any)
}

// outputWriter writes to stdout/stderr for normal CLI usage
type cliWriter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewWriter creates a console writer writer
func NewWriter() Writer {
	return &cliWriter{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Printf writes formatted writer to stdout
func (c *cliWriter) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(c.stdout, format, args...)
}

// Errorf writes formatted error writer to stderr
func (c *cliWriter) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(c.stderr, format, args...)
}

// Package-level output variables and synchronization
var (
	writer  Writer       // writer is the global output writer instance used for CLI operations
	printMu sync.RWMutex // synchronizes Printf access
	errorMu sync.RWMutex // synchronizes Errorf access
)

// SetWriter sets the global writer writer (primarily for testing)
func SetWriter(w Writer) {
	printMu.Lock()
	defer printMu.Unlock()
	writer = w
	ensureWriter()
}

// GetWriter returns the current writer writer
func GetWriter() Writer {
	printMu.RLock()
	defer printMu.RUnlock()
	return writer
}

// Package-level convenience functions

// Printf writes formatted writer
func Printf(format string, args ...any) {
	printMu.RLock()
	defer printMu.RUnlock()
	writer.Printf(format, args...)
}

// Errorf writes to formatted error writer
func Errorf(format string, args ...any) {
	errorMu.RLock()
	defer errorMu.RUnlock()
	writer.Errorf(format, args...)
}

// ensureWriter panics if no Writer has been set, preventing uninitialized usage
func ensureWriter() {
	if writer == nil {
		panic("Must set Writer with cliutil.SetWriter() before using cliutil package")
	}
}
