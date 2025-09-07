package common

import (
	"errors"
)

var (
	ErrPathIsDir        = errors.New("path is a directory")
	ErrFileDoesNotExist = errors.New("file does not exist")
	ErrFileExists       = errors.New("file exists")
)
