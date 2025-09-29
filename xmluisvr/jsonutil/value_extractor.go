package jsonutil

import (
	"bytes"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"io"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// ExtractValuesFromReader processes multiple selectors in a single pass through JSON.
// Returns values for found selectors, list of selectors that were found, and any errors.
// Continues processing all selectors even when some fail to provide comprehensive error reporting.
func ExtractValuesFromReader(reader io.Reader, selectors []common.Selector) (values []any, found []common.Selector, err error) {
	var buffer bytes.Buffer
	var teeReader io.Reader
	var errs []error
	var rawBytes []byte

	if reader == nil {
		err = errors.Join(
			ErrJSONPathTraversalFailed,
			ErrJSONBodyCannotBeEmpty,
			fmt.Errorf("selectors=%v", selectors),
		)
		goto end
	}

	if len(selectors) == 0 {
		err = errors.Join(
			ErrJSONPathTraversalFailed,
			ErrJSONValueSelectorCannotBeEmpty,
		)
		goto end
	}

	// Set up streaming with TeeReader to capture raw bytes
	teeReader = io.TeeReader(reader, &buffer)
	rawBytes, err = readAllBytes(teeReader)
	if err != nil {
		err = errors.Join(
			ErrJSONStreamingParseFailed,
			ErrJSONReadFailed,
			fmt.Errorf("error=%v", err),
		)
		goto end
	}

	values = make([]any, len(selectors))
	found = make([]common.Selector, 0, len(selectors))

	// Process each selector individually
	for i, selector := range selectors {
		var value any
		var selectorErr error

		// Create fresh reader for each selector
		selectorReader := bytes.NewReader(rawBytes)
		value, selectorErr = extractSingleValue(selectorReader, selector, rawBytes)
		if selectorErr != nil {
			errs = append(errs, selectorErr)
			continue
		}

		values[i] = value
		found = append(found, selector)
	}

	// Join all collected errors
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

end:
	return values, found, err
}

// ExtractValuesFromBytes is a convenience wrapper for ExtractValuesFromReader
func ExtractValuesFromBytes(jsonBytes []byte, selectors []common.Selector) (values []any, found []common.Selector, err error) {
	if len(jsonBytes) == 0 {
		err = errors.Join(
			ErrJSONPathTraversalFailed,
			ErrJSONBodyCannotBeEmpty,
			fmt.Errorf("selectors=%v", selectors),
		)
		goto end
	}

	values, found, err = ExtractValuesFromReader(bytes.NewReader(jsonBytes), selectors)

end:
	return values, found, err
}

// ExtractValueFromReader extracts a single value from JSON - convenience wrapper
func ExtractValueFromReader(reader io.Reader, selector common.Selector) (value any, err error) {
	var values []any
	var found []common.Selector

	values, found, err = ExtractValuesFromReader(reader, []common.Selector{selector})
	if err != nil {
		goto end
	}

	if len(found) == 0 {
		err = fmt.Errorf("selector_not_found=%s", selector)
		goto end
	}

	value = values[0]

end:
	return value, err
}

// ExtractValueFromBytes extracts a single value from JSON bytes - convenience wrapper
func ExtractValueFromBytes(jsonBytes []byte, selector common.Selector) (value any, err error) {
	var values []any
	var found []common.Selector

	values, found, err = ExtractValuesFromBytes(jsonBytes, []common.Selector{selector})
	if err != nil {
		goto end
	}

	if len(found) == 0 {
		err = fmt.Errorf("selector_not_found=%s", selector)
		goto end
	}

	value = values[0]

end:
	return value, err
}

// extractSingleValue handles extraction of a single selector from JSON
func extractSingleValue(reader io.Reader, selector common.Selector, rawBytes []byte) (value any, err error) {
	var decoder *jsontext.Decoder
	var state *extractState

	if len(selector) == 0 {
		err = errors.Join(
			ErrJSONPathTraversalFailed,
			ErrJSONValueSelectorCannotBeEmpty,
		)
		goto end
	}

	decoder = jsontext.NewDecoder(reader)
	state = newExtractState(decoder, string(selector), rawBytes)

	// Navigate through each path segment
	for i, segment := range state.segments {
		state.position = i
		if segment == "" {
			err = state.joinErrors(
				ErrJSONPathTraversalFailed,
				ErrJSONPathContainsEmptySegment,
			)
			goto end
		}

		err = state.navigateToSegment(segment)
		if err != nil {
			goto end
		}
		state.pathProgress = append(state.pathProgress, segment)
	}

	// Extract the final value
	err = jsonv2.UnmarshalDecode(decoder, &value)
	if err != nil {
		err = state.joinErrors(
			ErrJSONStreamingParseFailed,
			ErrJSONUnmarshalFailed,
			err,
		)
	}

end:
	return value, err
}

// readAllBytes reads all bytes from a reader
func readAllBytes(reader io.Reader) ([]byte, error) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(reader)
	return buffer.Bytes(), err
}
