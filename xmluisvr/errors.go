package xmluisvr

import (
	"errors"
)

var (
	// ErrPathIsDir indicates that a requested file path is actually a directory.
	ErrPathIsDir = errors.New("path is a directory")
)
