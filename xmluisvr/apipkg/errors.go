package apipkg

import (
	"errors"
)

var (
	ErrNoAPIProvided                = errors.New("no api provided")
	ErrInvalidRowsExpectedType      = errors.New("invalid rows expected type")
	ErrInvalidResultsColumnDataType = errors.New("invalid results column data type")
	ErrInvalidRowType               = errors.New("invalid row type")
	ErrInvalidDataType              = errors.New("invalid data type")
)
