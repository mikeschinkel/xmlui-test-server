package apipkg

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/jsonxtractr"
)

var (
	ErrExtractingFromReader     = jsonxtractr.ErrExtractingFromJSONByReader
	ErrExtractingJSONBodyValues = jsonxtractr.ErrExtractingJSONBodyValues
	ErrSelectorNotFound         = jsonxtractr.ErrJSONSelectorNotFound
)

var (
	// ErrInvalidAPIEndpointParameters indicates that the endpoint parameters configuration is invalid.
	ErrInvalidAPIEndpointParameters = errors.New("invalid API endpoint parameters")

	// ErrInvalidAPIEndpointParameter indicates that a specific parameter configuration is invalid.
	ErrInvalidAPIEndpointParameter = errors.New("invalid API endpoint parameter")

	ErrInvalidAPIEndpoint = errors.New("invalid API endpoint")

	// ErrCannotTypeAssert indicates a type assertion failure during parameter parsing.
	ErrCannotTypeAssert = errors.New("cannot type assert")

	// ErrOneOfQueryAndQueryFileMustNotBeEmpty is returned when an endpoint has neither
	// an inline query nor a query file specified.
	ErrOneOfQueryAndQueryFileMustNotBeEmpty = errors.New("at least one of query for query file must not be empty")

	ErrQueryTypeParsingNotYetSupported = errors.New("query type parsing not yet supported")

	ErrGettingResponseContent = errors.New("error getting response content")

	ErrNoResponsePayloadFound = errors.New("no response payload found")

	ErrRouteNotMatched = errors.New("route not matched")

	ErrRouteMatchingFailed = errors.New("route matching failed")

	ErrQueryValuesExtractionFailed = errors.New("query values extraction failed")
)
