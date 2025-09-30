package testutil

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
)

// BufferedWriter implements cliutil.Writer and captures all output in buffers for testing
type BufferedWriter struct {
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	mu         sync.RWMutex
	quiet      bool
	verbosity  int
	useLevel   int
	loudWriter cliutil.Writer
	v2Writer   cliutil.Writer
	v3Writer   cliutil.Writer
}

// Verify BufferedWriter implements cliutil.Writer interface
var _ cliutil.Writer = (*BufferedWriter)(nil)

// NewBufferedWriter creates a new BufferedWriter with default settings
func NewBufferedWriter() *BufferedWriter {
	return &BufferedWriter{
		stdout:    &bytes.Buffer{},
		stderr:    &bytes.Buffer{},
		quiet:     false,
		verbosity: 3, // Default to max verbosity for testing
		useLevel:  1, // Default level
	}
}

// NewBufferedWriterWithVerbosity creates a BufferedWriter with specified verbosity level
func NewBufferedWriterWithVerbosity(verbosity int) *BufferedWriter {
	if verbosity < 1 || verbosity > 3 {
		panic(fmt.Sprintf("Invalid verbosity for BufferedWriter; must be between 1-3; got %d", verbosity))
	}
	return &BufferedWriter{
		stdout:    &bytes.Buffer{},
		stderr:    &bytes.Buffer{},
		quiet:     false,
		verbosity: verbosity,
		useLevel:  1,
	}
}

// Printf writes formatted output to stdout buffer
func (w *BufferedWriter) Printf(format string, args ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.quiet {
		return
	}
	if w.verbosity < w.useLevel {
		return
	}

	formatted := fmt.Sprintf(format, args...)
	w.stdout.WriteString(formatted)
}

// Errorf writes formatted error output to stderr buffer
func (w *BufferedWriter) Errorf(format string, args ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Process error arguments to flatten newlines (same as cliutil.cliWriter)
	processedArgs := make([]any, len(args))
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			processedArgs[i] = strings.Replace(err.Error(), "\n", "; ", -1)
		} else {
			processedArgs[i] = arg
		}
	}

	formatted := fmt.Sprintf(format, processedArgs...)
	w.stderr.WriteString(formatted)
}

// Loud returns a Writer that ignores the quiet setting
func (w *BufferedWriter) Loud() cliutil.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.loudWriter != nil {
		return w.loudWriter
	}

	w.loudWriter = &BufferedWriter{
		stdout:    w.stdout, // Share the same buffers
		stderr:    w.stderr,
		quiet:     false, // Always loud
		verbosity: w.verbosity,
		useLevel:  w.useLevel,
	}
	return w.loudWriter
}

// V2 returns a Writer for verbosity level 2
func (w *BufferedWriter) V2() cliutil.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.v2Writer != nil {
		return w.v2Writer
	}

	w.v2Writer = &BufferedWriter{
		stdout:    w.stdout, // Share the same buffers
		stderr:    w.stderr,
		quiet:     w.quiet,
		verbosity: w.verbosity,
		useLevel:  2, // Level 2
	}
	return w.v2Writer
}

// V3 returns a Writer for verbosity level 3
func (w *BufferedWriter) V3() cliutil.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.v3Writer != nil {
		return w.v3Writer
	}

	w.v3Writer = &BufferedWriter{
		stdout:    w.stdout, // Share the same buffers
		stderr:    w.stderr,
		quiet:     w.quiet,
		verbosity: w.verbosity,
		useLevel:  3, // Level 3
	}
	return w.v3Writer
}

// Testing helper methods

// GetStdout returns the current stdout buffer contents
func (w *BufferedWriter) GetStdout() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stdout.String()
}

// GetStderr returns the current stderr buffer contents
func (w *BufferedWriter) GetStderr() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stderr.String()
}

// GetAllOutput returns both stdout and stderr combined
func (w *BufferedWriter) GetAllOutput() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.stdout.String() + w.stderr.String()
}

// ContainsStdout returns true if stdout buffer contains the specified substring
func (w *BufferedWriter) ContainsStdout(s string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return strings.Contains(w.stdout.String(), s)
}

// ContainsStderr returns true if stderr buffer contains the specified substring
func (w *BufferedWriter) ContainsStderr(s string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return strings.Contains(w.stderr.String(), s)
}

// ContainsOutput returns true if either buffer contains the specified substring
func (w *BufferedWriter) ContainsOutput(s string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return strings.Contains(w.stdout.String(), s) || strings.Contains(w.stderr.String(), s)
}

// Reset clears both stdout and stderr buffers
func (w *BufferedWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stdout.Reset()
	w.stderr.Reset()
}

// SetQuiet sets the quiet mode (suppresses all Printf output)
func (w *BufferedWriter) SetQuiet(quiet bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.quiet = quiet
}

// SetVerbosity sets the verbosity level (1-3)
func (w *BufferedWriter) SetVerbosity(verbosity int) {
	if verbosity < 1 || verbosity > 3 {
		panic(fmt.Sprintf("Invalid verbosity for BufferedWriter; must be between 1-3; got %d", verbosity))
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.verbosity = verbosity
}

// GetStdoutLines returns stdout content split into lines (excluding empty lines)
func (w *BufferedWriter) GetStdoutLines() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	content := w.stdout.String()
	if content == "" {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

// GetStderrLines returns stderr content split into lines (excluding empty lines)
func (w *BufferedWriter) GetStderrLines() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	content := w.stderr.String()
	if content == "" {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

// CountStdoutLines returns the number of non-empty lines in stdout
func (w *BufferedWriter) CountStdoutLines() int {
	return len(w.GetStdoutLines())
}

// CountStderrLines returns the number of non-empty lines in stderr
func (w *BufferedWriter) CountStderrLines() int {
	return len(w.GetStderrLines())
}
