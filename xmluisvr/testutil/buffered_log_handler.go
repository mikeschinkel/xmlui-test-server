package testutil

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// BufferedLogHandler implements slog.Handler and captures logs in a buffer
type BufferedLogHandler struct {
	opts   slog.HandlerOptions
	buffer *bytes.Buffer
	mu     sync.Mutex
}

// NewBufferedLogHandler creates a new BufferedLogHandler
func NewBufferedLogHandler() *BufferedLogHandler {
	return &BufferedLogHandler{
		buffer: &bytes.Buffer{},
	}
}

// Enabled implements slog.Handler
func (h *BufferedLogHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler
func (h *BufferedLogHandler) Handle(_ context.Context, r slog.Record) (err error) {
	var data []byte

	h.mu.Lock()
	defer h.mu.Unlock()

	entry := NewLogEntry(r)

	// Add attributes
	r.Attrs(func(attr slog.Attr) bool {
		entry.Attrs = append(entry.Attrs, attr.String())
		return true
	})

	// Write to buffer
	data, err = json.Marshal(entry)
	if err != nil {
		goto end
	}
	h.buffer.Write(data)
	h.buffer.WriteByte('\n')
end:
	return err
}

// WithAttrs implements slog.Handler
func (h *BufferedLogHandler) WithAttrs([]slog.Attr) slog.Handler {
	// For simplicity in testing, we don't need to implement this properly
	return h
}

// WithGroup implements slog.Handler
func (h *BufferedLogHandler) WithGroup(string) slog.Handler {
	// For simplicity in testing, we don't need to implement this properly
	return h
}

// Buffer returns the underlying buffer
func (h *BufferedLogHandler) Buffer() *bytes.Buffer {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buffer
}

// String returns the buffer contents as a string
func (h *BufferedLogHandler) String() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buffer.String()
}

// Reset clears the buffer
func (h *BufferedLogHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buffer.Reset()
}

// Contains returns true if the buffer contains the specified substring
func (h *BufferedLogHandler) Contains(s string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return bytes.Contains(h.buffer.Bytes(), []byte(s))
}

// GetLogEntries parses the buffer into log entries
func (h *BufferedLogHandler) GetLogEntries() ([]map[string]interface{}, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var entries []map[string]interface{}
	lines := bytes.Split(h.buffer.Bytes(), []byte("\n"))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var entry map[string]interface{}
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (h *BufferedLogHandler) GetLogEntriesByLevel(level slog.Level) (entries []LogEntry, err error) {
	// Declare all variables before first goto
	var scanner *bufio.Scanner
	var line []byte
	var entry LogEntry

	h.mu.Lock()
	defer h.mu.Unlock()

	// Create scanner to read buffer line by line
	scanner = bufio.NewScanner(bytes.NewReader(h.buffer.Bytes()))

	// Scan through each line
	for scanner.Scan() {
		line = scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Unmarshal JSON line into entry map
		entry = LogEntry{}
		err = json.Unmarshal(line, &entry)
		if err != nil {
			goto end
		}

		// Add entry to result if level matches
		if entry.Level == level.String() {
			entry.Level = "" // Caller knows what level it is
			entry.OmitDateTime = true
			entries = append(entries, entry)
		}
	}

	// Check for scanner errors
	err = scanner.Err()

end:
	return entries, err
}
