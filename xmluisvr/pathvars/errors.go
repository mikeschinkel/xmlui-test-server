package pathvars

import (
	"errors"
)

// Sentinel errors
var (
	ErrInvalidTemplate       = errors.New("invalid template syntax")
	ErrUnmatchedBrace        = errors.New("unmatched brace in template")
	ErrInvalidParameter      = errors.New("invalid parameter definition")
	ErrInvalidType           = errors.New("unknown parameter type")
	ErrInvalidConstraint     = errors.New("invalid constraint syntax")
	ErrNoMatch               = errors.New("no matching route")
	ErrAPIRouterNotCompiled  = errors.New("API router not compiled; must be compiled before calling Match()")
	ErrValidationFailed      = errors.New("parameter validation failed")
	ErrUnknownConstraintType = errors.New("unknown constraint type")
	ErrInvalidSyntax         = errors.New("invalid syntax")
	ErrParseFailed           = errors.New("parse failed")
)
