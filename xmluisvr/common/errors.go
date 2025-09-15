package common

import (
	"errors"
)

var (
	ErrPathIsDir        = errors.New("path is a directory")
	ErrFileDoesNotExist = errors.New("file does not exist")
	ErrFileExists       = errors.New("file exists")
)

var (
	ErrNoAPIProvided                = errors.New("no api provided")
	ErrInvalidCardinalityType       = errors.New("invalid cardinality type")
	ErrInvalidResultsColumnDataType = errors.New("invalid results column data type")
	ErrInvalidRowType               = errors.New("invalid row type")
	ErrInvalidDataType              = errors.New("invalid data type")
)
