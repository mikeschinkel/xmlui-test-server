package dbpkg

import (
	"errors"
)

var (
	ErrConnFailed                = errors.New("failed to connect to database")
	ErrInvalidConnString         = errors.New("invalid connection string")
	ErrConnectStringNotSupported = errors.New("connection string is not valid for any supported database")
)
