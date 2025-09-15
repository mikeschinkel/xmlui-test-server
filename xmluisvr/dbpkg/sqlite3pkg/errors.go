package sqlite3pkg

import (
	"errors"
)

var (
	ErrFailedToTypeAssertToSQLite3Extension = errors.New("failed to type assert to SQLite3 extension")
	ErrInEventQueryForSQLite3Extension      = errors.New("error in event query for SQLite3 extension")
	ErrInvalidWALAutocheckpointValue        = errors.New("invalid WAl autocheckpoint value")
	ErrInvalidBusyTimeoutValue              = errors.New("invalid busy timeout value")
)
