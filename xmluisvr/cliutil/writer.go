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
	Loud() Writer
	V2() Writer
	V3() Writer
}

var _ Writer = (*cliWriter)(nil)

// outputWriter writes to stdout/stderr for normal CLI usage
type cliWriter struct {
	stdout    io.Writer
	stderr    io.Writer
	quiet     bool
	loud      Writer
	v2        Writer
	v3        Writer
	useLevel  int
	verbosity int
}

func (c *cliWriter) V2() Writer {
	if c.v2 != nil {
		goto end
	}
	c.v2 = &cliWriter{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		verbosity: c.verbosity,
		useLevel:  2,
	}
end:
	return c.v2
}

func (c *cliWriter) V3() Writer {
	if c.v3 != nil {
		goto end
	}
	c.v3 = &cliWriter{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		verbosity: c.verbosity,
		useLevel:  3,
	}
end:
	return c.v3
}

func (c *cliWriter) Loud() Writer {
	if c.loud != nil {
		goto end
	}
	c.loud = &cliWriter{
		stdout: os.Stdout,
		stderr: os.Stderr,
		quiet:  false,
	}
end:
	return c.loud
}

type WriterArgs struct {
	Quiet     bool
	Verbosity int
}

// NewWriter creates a console writer writer
func NewWriter(args WriterArgs) Writer {
	if args.Verbosity < 1 || 3 < args.Verbosity {
		panic(fmt.Sprintf("Invalid verbosity for cliutil.Writer.SetVerbosity(); must be between 1-3; got %d", args.Verbosity))
	}
	return &cliWriter{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		quiet:     args.Quiet,
		verbosity: args.Verbosity,
	}
}

// Printf writes formatted writer to stdout
func (c *cliWriter) Printf(format string, args ...any) {
	if c.quiet {
		goto end
	}
	if c.verbosity < c.useLevel {
		goto end
	}
	_, _ = fmt.Fprintf(c.stdout, format, args...)
end:
	return
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
