// Package pathvars/errors defines error values used throughout the pathvars package.
// These sentinel errors provide specific error types for different failure modes
// during path template parsing, route compilation, and request matching.
package pathvars

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars/pvtypes"
)

// Sentinel errors for various pathvars operations.
var (

	// ErrInvalidParameter indicates that a parameter is invalid
	ErrInvalidParameter = pvtypes.ErrInvalidParameter

	// ErrInvalidParameterSyntax indicates that a parameter syntax is malformed.
	ErrInvalidParameterSyntax = pvtypes.ErrInvalidParameterSyntax

	// ErrInvalidParameterType indicates that an unknown or unsupported parameter type was specified.
	ErrInvalidParameterType = pvtypes.ErrInvalidParameterSyntax

	// ErrParameterValidationFailed indicates that parameter validation failed against its type or constraints.
	ErrParameterValidationFailed = pvtypes.ErrParameterValidationFailed

	// ErrParseFailed indicates that parsing of a constraint or parameter failed.
	ErrParseFailed = pvtypes.ErrParseFailed

	// ErrInvalidConstraint indicates that a parameter constraint has invalid syntax.
	ErrInvalidConstraint = pvtypes.ErrInvalidConstraint

	// ErrInvalidLengthRangeMinGreaterThanMax indicates that minimum length is greater than maximum.
	ErrInvalidLengthRangeMinGreaterThanMax = pvtypes.ErrInvalidLengthRangeMinGreaterThanMax

	// ErrInvalidLengthRangeNegativeMin indicates that minimum length is negative.
	ErrInvalidLengthRangeNegativeMin = pvtypes.ErrInvalidLengthRangeNegativeMin

	// ErrInvalidDateFormatSpec indicates that date format specification is invalid.
	ErrInvalidDateFormatSpec = pvtypes.ErrInvalidDateFormat

	ErrNameSpecNameCannotBeEmpty  = pvtypes.ErrNameSpecNameCannotBeEmpty
	ErrConstraintValidationFailed = pvtypes.ErrConstraintValidationFailed
)

// Sentinel errors for various pathvars operations.
var (
	// ErrInvalidTemplate indicates that a path template has invalid syntax.
	ErrInvalidTemplate = errors.New("invalid template syntax")

	ErrExpectedRangeFormat    = errors.New("expected format 'range[min..max]")
	ErrExpectedLengthFormat   = errors.New("expected format 'length['min..max]")
	ErrExpectedDateOnlyFormat = errors.New("expected YYYY-MM-DD format")

	// ErrNoMatch indicates that no route matched the incoming request.
	ErrNoMatch = errors.New("no matching route")

	ErrParsingDBExtensionFailed = errors.New("parsing database extension failed")

	ErrRequiredParameterNotProvided = errors.New("required parameter not provided")

	ErrInvalidURLQueryString = errors.New("invalid URL query string")

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

	// ErrInvalidCUIDFormat indicates that value is not a valid CUID.
	ErrInvalidCUIDFormat = errors.New("invalid CUID format")

	// ErrInvalidSnowflakeFormat indicates that value is not a valid Snowflake ID.
	ErrInvalidSnowflakeFormat = errors.New("invalid Snowflake ID format")

	// ErrSnowflakeTimestampInFuture indicates that Snowflake timestamp is in the future.
	ErrSnowflakeTimestampInFuture = errors.New("Snowflake timestamp is in the future")

	// ErrSnowflakeTimestampNegative indicates that Snowflake timestamp is negative (before epoch).
	ErrSnowflakeTimestampNegative = errors.New("Snowflake timestamp is negative (before epoch)")

	// ErrInvalidSnowflakeEpoch indicates that the custom epoch parameter is invalid.
	ErrInvalidSnowflakeEpoch = errors.New("invalid Snowflake epoch parameter")

	// Date Format Constraint Errors

	// ErrStringFormatOnlySupportsIDFormats indicates that string format constraint only supports ulid, ksuid, nanoid.
	ErrStringFormatOnlySupportsIDFormats = errors.New("string format constraint only supports ulid, ksuid, nanoid")

	// ErrFormatConstraintUnsupportedDataType indicates that format constraint only supports date, uuid, and string data types.
	ErrFormatConstraintUnsupportedDataType = errors.New("format constraint only supports date, uuid, and string data types")

	// ErrMoreSegmentsThanFormat indicates that value has more segments than format specification.
	ErrMoreSegmentsThanFormat = errors.New("more segments in value than in format")

	// ErrFailedToBuildPartialLayout indicates that building partial date layout failed.
	ErrFailedToBuildPartialLayout = errors.New("failed to build partial layout")

	// ErrNoValidDateTimeTokens indicates that no valid date/time tokens were found in format.
	ErrNoValidDateTimeTokens = errors.New("no valid date/time tokens found in format")

	// ErrAmbiguousMMToken indicates that 'mm' token is ambiguous.
	ErrAmbiguousMMToken = errors.New("ambiguous 'mm' token - use 'ii' for minutes or add other tokens for context")

	// Length Constraint Errors

	// Regex Constraint Errors

	// ErrEmptyRegexPattern indicates that regex pattern is empty.
	ErrEmptyRegexPattern = errors.New("empty regex pattern")

	// ErrInvalidRegexPattern indicates that regex pattern is invalid.
	ErrInvalidRegexPattern = errors.New("invalid regular expression")

	// ErrRegexPatternContainsStartAnchor indicates that regex pattern contains ^ start anchor.
	ErrRegexPatternContainsStartAnchor = errors.New("regex pattern contains ^ start anchor")

	// ErrRegexPatternContainsEndAnchor indicates that regex pattern contains $ end anchor.
	ErrRegexPatternContainsEndAnchor = errors.New("regex pattern contains $ end anchor")

	// ErrRegexPatternContainsBothAnchors indicates that regex pattern contains both ^ and $ anchors.
	ErrRegexPatternContainsBothAnchors = errors.New("regex pattern contains both ^ and $ anchors")

	ErrParameterNotFoundInValuesMap      = errors.New("parameter not found in values map")
	ErrPathParameterNotFoundInValuesMap  = errors.New("path parameter not found in values map")
	ErrQueryParameterNotFoundInValuesMap = errors.New("query parameter not found in values map")
)

var (
	ErrInvalidRangeConstraint = errors.New("invalid range constraint")
	ErrInvalidRegexConstraint = errors.New("invalid regex constraint")
	ErrInvalidEnumConstraint  = errors.New("invalid enum constraint")
)

var (
	ErrInvalidRangeValue   = errors.New("invalid range value")
	ErrInvalidMinimumValue = errors.New("invalid minimum value")
	ErrInvalidMaximumValue = errors.New("invalid maximum value")
	ErrInvalidMinMaxValue  = errors.New("minimum value cannot be less than maximum value")
	ErrInvalidMinMaxDate   = errors.New("minimum date cannot be less than maximum date")
	ErrInvalidDateFormat   = errors.New("invalid date format")
)
