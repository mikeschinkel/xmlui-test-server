package jsonxtractr

import (
	"bytes"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io"
)

type ValuesMap map[Selector]any

// ExtractValuesFromReader processes multiple selectors in a single pass through JSON.
// Returns values for found selectors, list of selectors that were found, and any errors.
// Continues processing all selectors even when some fail to provide comprehensive error reporting.
func ExtractValuesFromReader(reader io.Reader, selectors []Selector) (valuesMap ValuesMap, notFound []Selector, err error) {
	var buffer bytes.Buffer
	var teeReader io.Reader
	var errs []error
	var rawBytes []byte

	if reader == nil {
		err = NewErr(
			ErrJSONPathTraversalFailed,
			ErrJSONBodyCannotBeEmpty,
			"selectors", selectors,
		)
		goto end
	}

	if len(selectors) == 0 {
		err = NewErr(
			ErrJSONPathTraversalFailed,
			ErrJSONValueSelectorCannotBeEmpty,
		)
		goto end
	}

	// Set up streaming with TeeReader to capture raw bytes
	teeReader = io.TeeReader(reader, &buffer)
	rawBytes, err = readAllBytes(teeReader)
	if err != nil {
		err = NewErr(
			ErrJSONStreamingParseFailed,
			ErrJSONReadFailed,
			err,
		)
		goto end
	}

	valuesMap = make(ValuesMap, len(selectors))
	notFound = make([]Selector, 0, len(selectors))

	// Process each selector individually
	for _, selector := range selectors {
		var value any
		var selectorErr error

		// Create fresh reader for each selector
		selectorReader := bytes.NewReader(rawBytes)
		value, selectorErr = extractSingleValue(selectorReader, selector, rawBytes)
		if selectorErr != nil {
			errs = append(errs, selectorErr)
			continue
		}

		valuesMap[selector] = value
	}

	// Join all collected errors
	if len(errs) > 0 {
		err = CombineErrs(errs)
	}

	// Not create the list of selectors not found.
	for _, s := range selectors {
		_, ok := valuesMap[s]
		if ok {
			continue
		}
		notFound = append(notFound, s)
	}

end:
	return valuesMap, notFound, err
}

// ExtractValuesFromBytes is a convenience wrapper for ExtractValuesFromReader
func ExtractValuesFromBytes(jsonBytes []byte, selectors []Selector) (valuesMap ValuesMap, found []Selector, err error) {
	if len(jsonBytes) == 0 {
		err = NewErr(
			ErrJSONPathTraversalFailed,
			ErrJSONBodyCannotBeEmpty,
			"selectors", selectors,
		)
		goto end
	}

	valuesMap, found, err = ExtractValuesFromReader(bytes.NewReader(jsonBytes), selectors)

end:
	return valuesMap, found, err
}

// ExtractValueFromReader extracts a single value from JSON - convenience wrapper
func ExtractValueFromReader(reader io.Reader, selector Selector) (value any, err error) {
	var valuesMap ValuesMap
	var notFound []Selector
	var ok bool

	valuesMap, notFound, err = ExtractValuesFromReader(reader, []Selector{selector})
	if err != nil {
		err = WithErr(
			ErrFailedToExtractValueFromJSON,
			ErrExtractingFromJSONByReader,
			"selector", selector,
			err,
		)
		goto end
	}

	if len(notFound) > 0 {
		err = NewErr(
			ErrJSONSelectorNotFound,
			ErrExtractingFromJSONByReader,
			"selector", selector)
		goto end
	}

	value, ok = valuesMap[selector]
	if !ok {
		err = NewErr(
			ErrJSONSelectorNotFound,
			ErrExtractingFromJSONByReader,
			"selector", selector)
		goto end
	}

end:
	return value, err
}

// ExtractValueFromBytes extracts a single value from JSON bytes - convenience wrapper
func ExtractValueFromBytes(jsonBytes []byte, selector Selector) (value any, err error) {
	var valuesMap ValuesMap
	var notFound []Selector
	var ok bool

	valuesMap, notFound, err = ExtractValuesFromBytes(jsonBytes, []Selector{selector})
	if err != nil {
		err = WithErr(
			ErrFailedToExtractValueFromJSON,
			ErrExtractingFromJSONBytes,
			"selector", selector,
			err,
		)
		goto end
	}

	if len(notFound) > 0 {
		err = NewErr(
			ErrJSONSelectorNotFound,
			ErrExtractingFromJSONBytes,
			"selector", selector)
		goto end
	}

	value, ok = valuesMap[selector]
	if !ok {
		err = NewErr(
			ErrJSONSelectorNotFound,
			ErrExtractingFromJSONBytes,
			"selector", selector)
		goto end
	}

end:
	return value, err
}

// extractSingleValue handles extraction of a single selector from JSON
func extractSingleValue(reader io.Reader, selector Selector, rawBytes []byte) (value any, err error) {
	var decoder *jsontext.Decoder
	var state *extractState

	if len(selector) == 0 {
		err = NewErr(
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
			err = state.enrichError(
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
		err = state.enrichError(
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
