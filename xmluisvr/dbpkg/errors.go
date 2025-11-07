package dbpkg

import (
	"errors"
)

var (
	ErrConnectFailed                     = errors.New("failed to connect to database")
	ErrConnectStringNotSupported         = errors.New("connection string is not valid for any supported database")
	ErrInvalidConnectString              = errors.New("invalid connection string")
	ErrFailedToTypeAssertToExtensionType = errors.New("failed to type assert to expected type for database extension")
	ErrUnsupportedDBType                 = errors.New("unsupported database type")
	ErrExtensionsUnsupportedForDBType    = errors.New("extensions unsupported for database type")
)

var (
	ErrInvalidCardinality           = errors.New("invalid cardinality")
	ErrOneRowExpectedZeroReturned   = errors.New("one row expected but zero rows returned")
	ErrOneRowExpectedManyReturned   = errors.New("one row expected but many rows returned")
	ErrManyRowsExpectedZeroReturned = errors.New("many rows expected but zero rows returned")
)

var (
	ErrFailedToPingDatabase   = errors.New("failed to ping database")
	ErrFailedToOpenDatabase   = errors.New("failed to open database")
	ErrFailedToExecuteQueries = errors.New("failed to execute query(s)")
)
