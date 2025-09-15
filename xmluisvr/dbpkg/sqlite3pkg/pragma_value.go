package sqlite3pkg

import (
	"strconv"
	"time"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

const (
	JournalModePragma       common.Identifier = "journal_mode"
	SynchronousPragma       common.Identifier = "synchronous"
	ForeignKeysPragma       common.Identifier = "foreign_keys"
	BusyTimeoutPragma       common.Identifier = "busy_timeout"
	WALAutoCheckpointPragma common.Identifier = "wal_autocheckpoint"
)

var GetPragmasFuncs = map[common.Identifier]func(s *SQLite3) string{
	JournalModePragma:       func(s *SQLite3) string { return string(s.JournalMode) },
	SynchronousPragma:       func(s *SQLite3) string { return string(s.Synchronous) },
	ForeignKeysPragma:       func(s *SQLite3) string { return string(s.ForeignKeyMode) },
	BusyTimeoutPragma:       func(s *SQLite3) string { return strconv.FormatInt(s.BusyTimeout.Milliseconds(), 10) },
	WALAutoCheckpointPragma: func(s *SQLite3) string { return strconv.Itoa(s.AutoCheckpoint) },
}

const (
	DefaultJournalMode       = WALJournalMode
	DefaultSynchronous       = NormalSynchronous
	DefaultForeignKeyMode    = IgnoreForeignKeys
	DefaultWALAutoCheckpoint = 1000
)

var DefaultBusyTimeout = 5 * time.Second / time.Millisecond
