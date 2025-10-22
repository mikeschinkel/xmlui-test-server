package apiresp

import (
	"errors"
)

var (
	ErrUnspecifiedError             = errors.New("unspecified error")
	ErrFailedToReadHTTPRequestBody  = errors.New("failed to read HTTP request body")
	ErrNeitherDBQueryNorEndpointSet = errors.New("neither DBQuery nor Endpoint set")
	ErrEndpointParsedQueryNotSet    = errors.New("endpoint parsed query not set")
)

var (
	ErrNotUsingClonedPayloadArgs         = errors.New("attempting to mark property as being used by calling its Getter() but calling it on the original PayloadArgs and not the cloned version is invalid")
	ErrZeroValueForExpectedProperty      = errors.New("zero value for expected property")
	ErrNonZeroValueForUnexpectedProperty = errors.New("non-zero value for unexpected property")
	ErrInvalidPropertyUsage              = errors.New("invalid property usage")
)
