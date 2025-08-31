package dbutil

import (
	"errors"
)

var (
	ErrConnFailed = errors.New("failed to connect to database")
)
