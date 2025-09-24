package apipkg

import (
	"errors"
)

var (
	// ErrInvalidAPIEndpointParameters indicates that the endpoint parameters configuration is invalid.
	ErrInvalidAPIEndpointParameters = errors.New("invalid API endpoint parameters")

	// ErrInvalidAPIEndpointParameter indicates that a specific parameter configuration is invalid.
	ErrInvalidAPIEndpointParameter = errors.New("invalid API endpoint parameter")

	// ErrCannotTypeAssert indicates a type assertion failure during parameter parsing.
	ErrCannotTypeAssert = errors.New("cannot type assert")
)
