package dbpkg

import (
	"errors"
)

var (
	ErrConnFailed                        = errors.New("failed to connect to database")
	ErrInvalidConnString                 = errors.New("invalid connection string")
	ErrConnectStringNotSupported         = errors.New("connection string is not valid for any supported database")
	ErrFailedToTypeAssertToExtensionType = errors.New("failed to type assert to expected type for database extension")
	ErrUnsupportedDBType                 = errors.New("unsupported database type")
	ErrExtensionsUnsupportedForDBType    = errors.New("extensions unsupported for database type")
)
