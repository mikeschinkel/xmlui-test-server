package apiutil

import (
	"errors"
)

var (
	ErrUnspecifiedError             = errors.New("unspecified error")
	ErrFailedToReadHTTPRequestBody  = errors.New("failed to read HTTP request body")
	ErrNeitherDBQueryNorEndpointSet = errors.New("neither DBQuery nor Endpoint set")
	ErrEndpointParsedQueryNotSet    = errors.New("endpoint parsed query not set")
)
