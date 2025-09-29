// Package dbqvars/errors defines error values used throughout the dbqvars package.
// These sentinel errors provide specific error types for different failure modes
// during SQL parsing and placeholder processing.
package dbqvars

import (
	"errors"
)

// Sentinel errors for various dbqvars operations.
var (
	// ErrFormatParamFuncRequired indicates that ParseSQLArgs.FormatParamFunc is nil.
	ErrFormatParamFuncRequired = errors.New("ParseSQLArgs.GetFormatParamFunc is required")

	// ErrUnclosedPlaceholder indicates that a '{' placeholder was not closed with '}'.
	ErrUnclosedPlaceholder = errors.New("unclosed placeholder")

	// ErrInvalidPlaceholderName indicates that a placeholder name is invalid or malformed.
	ErrInvalidPlaceholderName = errors.New("invalid placeholder name")
)
