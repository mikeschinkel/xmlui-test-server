package cfgldr

import (
	"errors"
)

var (
	ErrAPIEndpointHasNoQueryOrFile = errors.New("APIConfig endpoint has no query or query file")
	ErrAPIEndpointMustNotBeEmpty   = errors.New("APIConfig endpoint must not be empty")
	ErrInvalidAPIEndpoint          = errors.New("invalid APIConfig endpoint")
	ErrInvalidAPIEndpointMethod    = errors.New("invalid APIConfig endpoint HTTP method")
	ErrInvalidAPIEndpointPath      = errors.New("invalid APIConfig endpoint path")
	ErrParseFailed                 = errors.New("parse failed")
	ErrReadFailed                  = errors.New("read failed")
)
