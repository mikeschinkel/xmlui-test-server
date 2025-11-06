package xmluisvr

import (
	"errors"
	"fmt"
)

var (
	// ErrPathIsDir indicates that a requested file path is actually a directory.
	ErrPathIsDir = errors.New("path is a directory")

	// ErrServerError indicates that the server terminated with an error condition.
	ErrServerError = fmt.Errorf("server terminated with an error")

	ErrUnauthorizedEndpointAccess = fmt.Errorf("attempt to access API endpoint without proper authorization")

	ErrNoConfigProvided    = errors.New("no config provided")
	ErrConfigParsingFailed = errors.New("config parsing failed")
)

var ErrInvalidServerPort = errors.New("invalid server port")

var (
	ErrFailedToParseDatabaseConfig = errors.New("failed to parse database configuration")
	ErrFailedToParseServerConfig   = errors.New("failed to parse server configuration")
	ErrFailedToParseAPIConfig      = errors.New("failed to parse API configuration")
)
