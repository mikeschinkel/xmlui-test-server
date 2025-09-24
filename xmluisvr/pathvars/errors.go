// Package pathvars/errors defines error values used throughout the pathvars package.
// These sentinel errors provide specific error types for different failure modes
// during path template parsing, route compilation, and request matching.
package pathvars

import (
	"errors"
)

// Sentinel errors for various pathvars operations.
var (
	// ErrInvalidTemplate indicates that a path template has invalid syntax.
	ErrInvalidTemplate = errors.New("invalid template syntax")

	// ErrUnmatchedBrace indicates that a template contains unmatched braces.
	ErrUnmatchedBrace = errors.New("unmatched brace in template")

	// ErrInvalidParameter indicates that a parameter definition is malformed.
	ErrInvalidParameter = errors.New("invalid parameter definition")

	// ErrInvalidParameterType indicates that an unknown or unsupported parameter type was specified.
	ErrInvalidParameterType = errors.New("unknown parameter type")

	// ErrInvalidConstraint indicates that a parameter constraint has invalid syntax.
	ErrInvalidConstraint = errors.New("invalid constraint syntax")

	// ErrNoMatch indicates that no route matched the incoming request.
	ErrNoMatch = errors.New("no matching route")

	// ErrAPIRouterNotCompiled indicates that Match() was called on an uncompiled router.
	ErrAPIRouterNotCompiled = errors.New("API router not compiled; must be compiled before calling Match()")

	// ErrValidationFailed indicates that parameter validation failed against its type or constraints.
	ErrValidationFailed = errors.New("parameter validation failed")

	// ErrUnknownConstraintType indicates that an unknown constraint type was specified.
	ErrUnknownConstraintType = errors.New("unknown constraint type")

	// ErrInvalidSyntax indicates that constraint or parameter syntax is malformed.
	ErrInvalidSyntax = errors.New("invalid syntax")

	// ErrParseFailed indicates that parsing of a constraint or parameter failed.
	ErrParseFailed = errors.New("parse failed")
)
