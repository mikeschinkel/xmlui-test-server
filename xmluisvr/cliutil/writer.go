// Package cliutil provides output management and synchronized writing for CLI applications.
package cliutil

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Writer defines the interface for user-facing writer
type Writer interface {
	Printf(string, ...any)
	Errorf(string, ...any)
	Quiet() bool
	SetQuiet(verbose bool)
	Loud() Writer
}

var _ Writer = (*cliWriter)(nil)

// outputWriter writes to stdout/stderr for normal CLI usage
type cliWriter struct {
	stdout io.Writer
	stderr io.Writer
	quiet  bool
	loud   *loudWriter
}
type loudWriter struct {
	cliWriter
}

func (c *cliWriter) Loud() Writer {
	if c.loud != nil {
		goto end
	}
	c.loud = &loudWriter{
		cliWriter: cliWriter{
			stdout: os.Stdout,
			stderr: os.Stderr,
			quiet:  false,
		},
	}
end:
	return c.loud
}

func (c *cliWriter) Quiet() bool {
	return c.quiet
}

func (c *cliWriter) SetQuiet(quiet bool) {
	c.quiet = quiet
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
	if c.quiet {
		return
	}
	_, _ = fmt.Fprintf(c.stdout, format, args...)
}

// Errorf writes formatted error writer to stderr
func (c *cliWriter) Errorf(format string, args ...any) {
	for i, arg := range args {
		err, ok := arg.(error)
		if !ok {
			continue
		}
		// Replace newlines in errors with semicolons
		args[i] = strings.Replace(err.Error(), "\n", "; ", -1)
	}
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

// Loud returns a Writer that ignores Quiet setting
func Loud() Writer {
	return writer.Loud()
}

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
