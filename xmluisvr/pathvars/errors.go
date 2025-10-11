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

	// ErrInvalidParameter indicates that a parameter is invalid
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrInvalidParameterSyntax indicates that a parameter syntax is malformed.
	ErrInvalidParameterSyntax = errors.New("invalid parameter syntax")

	// ErrInvalidParameterValue indicates that a parameter's value is invalid
	ErrInvalidParameterValue = errors.New("invalid parameter value")

	// ErrInvalidParameterType indicates that an unknown or unsupported parameter type was specified.
	ErrInvalidParameterType = errors.New("unknown parameter type")

	// ErrInvalidConstraint indicates that a parameter constraint has invalid syntax.
	ErrInvalidConstraint = errors.New("invalid constraint syntax")

	ErrExpectedRangeFormat       = errors.New("expected format 'range[min..max]")
	ErrExpectedLengthFormat      = errors.New("expected format 'length['min..max]")
	ErrExpectedISO8601DateFormat = errors.New("expected YYYY-MM-DD format")

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

	ErrInvalidNameSpec = errors.New("invalid name spec")

	ErrNameSpecNameCannotBeEmpty = errors.New("name spec name cannot be empty")

	ErrValueCannotBeEmpty = errors.New("value cannot be empty")

	ErrWhatNameSpecMustContain = errors.New("name spec must begin with a valid identifier — containing letters, digits and/or underscores — may then optionally contain an asterisk ('*': for multisegment), a question mark ('?': for optional), and if optional then optionally a default value, e.g. 'category?uncategorized'")

	ErrRequiredParameterNotProvided = errors.New("required parameter not provided")

	ErrInvalidURLQueryString = errors.New("invalid URL query string")

	ErrConstraintValidationFailed = errors.New("data type validation failed")

	ErrInvalidDataType = errors.New("invalid data type")

	ErrEnumValueIsEmpty = errors.New("enum constraint value is empty")

	ErrDateLessThanMinimum = errors.New("date must be greater than or equal to minimum")

	ErrDateGreaterThanMaximum = errors.New("date must be less than or equal to maximum")

	// Template/Parser Errors

	// ErrEmptyTemplate indicates that the template string is empty.
	ErrEmptyTemplate = errors.New("empty template")

	// ErrUnmatchedClosingBrace indicates an unmatched closing brace in template.
	ErrUnmatchedClosingBrace = errors.New("unmatched closing brace")

	// ErrUnmatchedOpeningBrace indicates an unmatched opening brace in template.
	ErrUnmatchedOpeningBrace = errors.New("unmatched opening brace(s)")

	// ErrMalformedBraces indicates a closing brace before and opening brace
	ErrMalformedBraces = errors.New("malformed brace; '{' must precede '}'")

	// Router Errors

	// ErrRouterNotCompiled indicates that the router must be compiled before matching.
	ErrRouterNotCompiled = errors.New("router must be compiled before matching")

	// ErrNoRouteMatched indicates that no route matched the request.
	ErrNoRouteMatched = errors.New("no route matched the request")

	// Parameter Errors

	// ErrParameterLocationNotSpecified indicates that parameter location was not specified.
	ErrParameterLocationNotSpecified = errors.New("parameter location not specified")

	// ErrDefaultValueValidationFailed indicates that default value validation failed.
	ErrDefaultValueValidationFailed = errors.New("default value validation failed")

	// ErrDefaultValueConstraintValidationFailed indicates that default value constraint validation failed.
	ErrDefaultValueConstraintValidationFailed = errors.New("default value constraint validation failed")

	// Data Type Validation Errors

	// ErrInvalidIntegerFormat indicates that value is not a valid integer.
	ErrInvalidIntegerFormat = errors.New("invalid integer format")

	// ErrInvalidDecimalFormat indicates that value is not a valid decimal.
	ErrInvalidDecimalFormat = errors.New("invalid decimal format")

	// ErrInvalidRealFormat indicates that value is not a valid real number.
	ErrInvalidRealFormat = errors.New("invalid real number format")

	// ErrUnsupportedDataType indicates that the data type is not supported.
	ErrUnsupportedDataType = errors.New("unsupported data type")

	// ErrInvalidIdentifierFormat indicates that value does not conform to identifier format.
	ErrInvalidIdentifierFormat = errors.New("must start with lowercase letter, followed by lowercase letters, digits, or underscores")

	// ErrDateValueEmpty indicates that date value cannot be empty.
	ErrDateValueEmpty = errors.New("date value cannot be empty")

	// ErrInvalidUUIDFormatBasic indicates that value is not a valid UUID (basic validation).
	ErrInvalidUUIDFormatBasic = errors.New("invalid UUID format (expected 8-4-4-4-12 hex digits)")

	// ErrInvalidAlphanumericFormat indicates that value must contain only letters and digits.
	ErrInvalidAlphanumericFormat = errors.New("value must contain only letters and digits")

	// ErrInvalidSlugFormat indicates that value does not conform to slug format.
	ErrInvalidSlugFormat = errors.New("must be lowercase letters/digits with optional hyphens between segments")

	// ErrInvalidBooleanFormat indicates that boolean value must be 'true' or 'false'.
	ErrInvalidBooleanFormat = errors.New("boolean value must be exactly 'true' or 'false'")

	// ErrInvalidEmailFormat indicates that value is not a valid email format.
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// UUID Constraint Errors

	// ErrUnsupportedUUIDFormat indicates that the UUID format is not supported.
	ErrUnsupportedUUIDFormat = errors.New("unsupported UUID format")

	// ErrUUIDVersionOutOfRange1to8 indicates that UUID version must be 1-8.
	ErrUUIDVersionOutOfRange1to8 = errors.New("UUID version must be 1-8")

	// ErrUUIDVersionOutOfRange1to5 indicates that UUID version must be 1-5.
	ErrUUIDVersionOutOfRange1to5 = errors.New("UUID version must be 1-5")

	// ErrUUIDVersionOutOfRange6to8 indicates that UUID version must be 6-8.
	ErrUUIDVersionOutOfRange6to8 = errors.New("UUID version must be 6-8")

	// ErrUUIDVersionMismatch indicates that UUID version does not match expected version.
	ErrUUIDVersionMismatch = errors.New("UUID version mismatch")

	// ErrInvalidUUIDShape indicates that UUID does not have the expected 8-4-4-4-12 format.
	ErrInvalidUUIDShape = errors.New("invalid UUID shape (expected 8-4-4-4-12 format)")

	// ErrInvalidUUIDHexEncoding indicates that UUID contains invalid hex encoding.
	ErrInvalidUUIDHexEncoding = errors.New("invalid hex encoding in UUID")

	// ErrInvalidUUIDVariant indicates that UUID variant is invalid (must be RFC 4122/9562).
	ErrInvalidUUIDVariant = errors.New("invalid UUID variant (must be RFC 4122/9562)")

	// ErrInvalidUUIDVersion indicates that UUID version is invalid (must be 1-8).
	ErrInvalidUUIDVersion = errors.New("invalid UUID version (must be 1-8)")

	// ErrInvalidULIDFormat indicates that value is not a valid ULID.
	ErrInvalidULIDFormat = errors.New("invalid ULID format")

	// ErrInvalidKSUIDFormat indicates that value is not a valid KSUID.
	ErrInvalidKSUIDFormat = errors.New("invalid KSUID format")

	// ErrInvalidNanoIDFormat indicates that value is not a valid NanoID.
	ErrInvalidNanoIDFormat = errors.New("invalid NanoID format")

	// Date Format Constraint Errors

	// ErrStringFormatOnlySupportsIDFormats indicates that string format constraint only supports ulid, ksuid, nanoid.
	ErrStringFormatOnlySupportsIDFormats = errors.New("string format constraint only supports ulid, ksuid, nanoid")

	// ErrFormatConstraintUnsupportedDataType indicates that format constraint only supports date, uuid, and string data types.
	ErrFormatConstraintUnsupportedDataType = errors.New("format constraint only supports date, uuid, and string data types")

	// ErrMoreSegmentsThanFormat indicates that value has more segments than format specification.
	ErrMoreSegmentsThanFormat = errors.New("more segments in value than in format")

	// ErrFailedToBuildPartialLayout indicates that building partial date layout failed.
	ErrFailedToBuildPartialLayout = errors.New("failed to build partial layout")

	// ErrInvalidDateFormatSpec indicates that date format specification is invalid.
	ErrInvalidDateFormatSpec = errors.New("invalid date format specification")

	// ErrNoValidDateTimeTokens indicates that no valid date/time tokens were found in format.
	ErrNoValidDateTimeTokens = errors.New("no valid date/time tokens found in format")

	// ErrAmbiguousMMToken indicates that 'mm' token is ambiguous.
	ErrAmbiguousMMToken = errors.New("ambiguous 'mm' token - use 'ii' for minutes or add other tokens for context")

	// Length Constraint Errors

	// ErrInvalidLengthRangeMinGreaterThanMax indicates that minimum length is greater than maximum.
	ErrInvalidLengthRangeMinGreaterThanMax = errors.New("invalid length range (min > max)")

	// ErrInvalidLengthRangeNegativeMin indicates that minimum length is negative.
	ErrInvalidLengthRangeNegativeMin = errors.New("invalid length range (min < 0)")

	// Regex Constraint Errors

	// ErrEmptyRegexPattern indicates that regex pattern is empty.
	ErrEmptyRegexPattern = errors.New("empty regex pattern")

	// ErrInvalidRegexPattern indicates that regex pattern is invalid.
	ErrInvalidRegexPattern = errors.New("invalid regular expression")

	// Constraint Type Errors

	// ErrInvalidConstraintTypeCharacter indicates that a constraint type contains an invalid character.
	ErrInvalidConstraintTypeCharacter = errors.New("invalid constraint type character")

	ErrParameterNotFoundInValuesMap = errors.New("parameter not found in values map")
)
