package cfgldr

import (
	"errors"
)

var (
	ErrReadFailed  = errors.New("read failed")
	ErrParseFailed = errors.New("parse failed")
)
