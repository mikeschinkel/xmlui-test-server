package testutil

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type LogEntries []LogEntry

func (ee LogEntries) String() string {
	var sb strings.Builder
	for i, entry := range ee {
		sb.WriteString(entry.String())
		if i == len(entry.Attrs)-1 {
			break
		}
		sb.WriteByte(';')
	}
	return sb.String()
}

type LogEntry struct {
	Level        string   `json:"level,omitempty"`
	Message      string   `json:"message"`
	DateTime     string   `json:"datetime,omitempty"`
	Attrs        []string `json:"attrs,omitempty"`
	OmitDateTime bool     `json:"-"`
}

func NewLogEntry(r slog.Record) *LogEntry {
	return &LogEntry{
		Level:    r.Level.String(),
		Message:  r.Message,
		DateTime: r.Time.Format(time.DateTime),
	}
}

func (e LogEntry) AttrsString() string {
	var sb strings.Builder
	for i, attr := range e.Attrs {
		sb.WriteString(attr)
		if i == len(e.Attrs)-1 {
			break
		}
		sb.WriteByte(' ')
	}
	return sb.String()
}

func (e LogEntry) String() (msg string) {
	msg = fmt.Sprintf("%s: %s", e.Level, e.Message)
	dt := ""
	if !e.OmitDateTime {
		dt = " at " + e.DateTime
	}
	return fmt.Sprintf("%s%s [%s]", msg, dt, e.AttrsString())
}
