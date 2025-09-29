package jsonutil

import (
	"errors"
)

// Sentinel errors for various jsonutil operations.
var (
	ErrJSONBodyCannotBeEmpty           = errors.New("JSON body cannot be empty")
	ErrJSONIndexOutOfRange             = errors.New("JSON index out of range")
	ErrJSONPathContainsEmptySegment    = errors.New("JSON path contains empty segment")
	ErrJSONPathExpectedArrayAtSegment  = errors.New("JSON path expected array at segment")
	ErrJSONPathExpectedObjectAtSegment = errors.New("JSON path expected object at segment")
	ErrJSONPathSegmentNotFound         = errors.New("JSON path segment not found")
	ErrJSONPathTraversalFailed         = errors.New("JSON path traversal failed")
	ErrJSONReadFailed                  = errors.New("JSON read failed")
	ErrJSONStreamingParseFailed        = errors.New("JSON streaming parse failed")
	ErrJSONTokenReadFailed             = errors.New("JSON token read failed")
	ErrJSONUnmarshalFailed             = errors.New("JSON unmarshal failed")
	ErrJSONValueSelectorCannotBeEmpty  = errors.New("JSON value selector cannot be empty")
)
