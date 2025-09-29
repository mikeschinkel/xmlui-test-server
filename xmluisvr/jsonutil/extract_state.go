package jsonutil

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type extractState struct {
	decoder      *jsontext.Decoder
	selector     string
	segments     []string
	pathProgress []string
	position     int
	rawBytes     []byte
}

func newExtractState(decoder *jsontext.Decoder, selector string, rawBytes []byte) *extractState {
	return &extractState{
		decoder:      decoder,
		selector:     selector,
		segments:     strings.Split(selector, "."),
		pathProgress: make([]string, 0),
		position:     0,
		rawBytes:     rawBytes,
	}
}

// navigateToSegment handles navigation to a specific segment in the JSON path
func (s *extractState) navigateToSegment(segment string) (err error) {

	// Check if this is a numeric index (array access)
	idx, parseErr := strconv.Atoi(segment)
	if parseErr == nil {
		err = s.navigateArrayIndex(idx, segment)
		goto end
	}

	// Handle object key access
	err = s.navigateObjectKey(segment)
end:
	return err
}

// navigateArrayIndex handles array index navigation
func (s *extractState) navigateArrayIndex(targetIdx int, segment string) (err error) {
	var currentIdx int
	var kind jsontext.Kind = s.decoder.PeekKind()

	// Check for negative index
	if targetIdx < 0 {
		err = s.joinErrors(nil,
			ErrJSONPathTraversalFailed,
			ErrJSONIndexOutOfRange,
			fmt.Errorf("target_index=%d", targetIdx),
		)
		goto end
	}

	if kind != '[' {
		err = s.joinErrors(nil,
			ErrJSONPathTraversalFailed,
			ErrJSONPathExpectedArrayAtSegment,
			fmt.Errorf("expected_type=%s", "array"),
			fmt.Errorf("actual_type=%s", kind.String()),
		)
		goto end
	}

	// Read array start token '['
	_, err = s.decoder.ReadToken()
	if err != nil {
		err = s.joinErrors(err,
			ErrJSONPathTraversalFailed,
			ErrJSONTokenReadFailed,
			fmt.Errorf("expected_token=%s", "array_start"),
		)
		goto end
	}

	// Skip elements until we reach the target index
	currentIdx = 0
	for currentIdx < targetIdx {
		if s.decoder.PeekKind() == ']' {
			err = s.joinErrors(nil,
				ErrJSONPathTraversalFailed,
				ErrJSONIndexOutOfRange,
				fmt.Errorf("target_index=%d", targetIdx),
				fmt.Errorf("array_length=%d", currentIdx),
			)
			goto end
		}
		err = s.decoder.SkipValue()
		if err != nil {
			err = s.joinErrors(err,
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				fmt.Errorf("skip_index=%d", currentIdx),
			)
			goto end
		}
		currentIdx++
	}

	// Check if we're at the end of array before target index
	if s.decoder.PeekKind() == ']' {
		err = s.joinErrors(nil,
			ErrJSONPathTraversalFailed,
			ErrJSONIndexOutOfRange,
			fmt.Errorf("target_index=%d", targetIdx),
			fmt.Errorf("array_length=%d", currentIdx),
		)
		goto end
	}
end:
	return err
}

// navigateObjectKey handles object key navigation
func (s *extractState) navigateObjectKey(targetKey string) (err error) {
	var availableKeys []string
	var keyToken jsontext.Token
	var kind jsontext.Kind = s.decoder.PeekKind()

	if kind != '{' {
		err = s.joinErrors(nil,
			ErrJSONPathTraversalFailed,
			ErrJSONPathExpectedObjectAtSegment,
			fmt.Errorf("expected_type=%s", "object"),
			fmt.Errorf("actual_type=%s", kind.String()),
		)
		goto end
	}

	// Read object start token '{'
	_, err = s.decoder.ReadToken()
	if err != nil {
		err = s.joinErrors(err,
			ErrJSONPathTraversalFailed,
			ErrJSONTokenReadFailed,
			fmt.Errorf("expected_token=%s", "object_start"),
		)
		goto end
	}

	// Collect available keys for error context
	availableKeys = make([]string, 0)

	// Search for the target key
	for s.decoder.PeekKind() != '}' {
		// Read the key
		keyToken, err = s.decoder.ReadToken()
		if err != nil {
			err = s.joinErrors(err,
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				fmt.Errorf("reading=%s", "object_key"),
			)
			goto end
		}

		key := keyToken.String()
		// Remove quotes from key
		if len(key) >= 2 && key[0] == '"' && key[len(key)-1] == '"' {
			key = key[1 : len(key)-1]
		}
		availableKeys = append(availableKeys, key)

		if key == targetKey {
			// Found the target key, the value is next
			goto end
		}

		// Skip the value for this key
		err = s.decoder.SkipValue()
		if err != nil {
			err = s.joinErrors(err,
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				fmt.Errorf("skipping_key=%s", key),
			)
			goto end
		}
	}

	// Key not found
	err = s.joinErrors(nil,
		ErrJSONPathTraversalFailed,
		ErrJSONPathSegmentNotFound,
		fmt.Errorf("missing_key=%s", targetKey),
		fmt.Errorf("available_keys=%v", availableKeys),
	)
end:
	return err
}

func (s *extractState) removeKeyQuotes(key string) string {
	if len(key) <= 1 {
		goto end
	}
	if key[0] != '"' {
		goto end
	}
	if key[len(key)-1] != '"' {
		goto end
	}
	key = key[1 : len(key)-1]
end:
	return key
}

// condensedJSON formats JSON in an easily comprehensible way
// that helps developers quickly locate and fix API configuration errors
func (s *extractState) condensedJSON() string {
	var formatted string
	var jsonStr string

	if len(s.rawBytes) == 0 {
		formatted = "JSON not available"
		goto end
	}

	jsonStr = string(s.rawBytes)

	// For empty or very short JSON, return as-is
	if len(jsonStr) <= 100 {
		formatted = jsonStr
		goto end
	}

	// For longer JSON, provide compact but readable format
	// Remove excessive whitespace while preserving structure
	formatted = strings.ReplaceAll(jsonStr, "\n", " ")
	formatted = strings.ReplaceAll(formatted, "\t", " ")
	// Collapse multiple spaces to single space
	for strings.Contains(formatted, "  ") {
		formatted = strings.ReplaceAll(formatted, "  ", " ")
	}

	// If still too long, intelligently truncate at JSON boundaries
	if len(formatted) > 200 {
		formatted = s.truncateAtJSONBoundary(formatted, 200)
	}

end:
	return formatted
}

// truncateAtJSONBoundary truncates at logical JSON structure points
func (s *extractState) truncateAtJSONBoundary(jsonStr string, maxLen int) string {
	var result string
	var truncated string
	var lastComma, lastBrace, lastBracket, cutPoint int

	if len(jsonStr) <= maxLen {
		result = jsonStr
		goto end
	}

	// Try to truncate at object or array boundaries for readability
	truncated = jsonStr[:maxLen-10] // Leave room for "...[more]"

	// Find last complete JSON structure
	lastComma = strings.LastIndex(truncated, ",")
	lastBrace = strings.LastIndex(truncated, "}")
	lastBracket = strings.LastIndex(truncated, "]")

	cutPoint = lastComma
	if lastBrace > cutPoint {
		cutPoint = lastBrace + 1
	}
	if lastBracket > cutPoint {
		cutPoint = lastBracket + 1
	}

	if cutPoint > 50 { // Ensure we don't cut too early
		result = jsonStr[:cutPoint] + "...[more]"
		goto end
	}

	// Fallback to simple truncation
	result = jsonStr[:maxLen-10] + "...[more]"

end:
	return result
}

// joinErrors expects an original error (or nil) then a list of other errors,
// adds state-specific errors of its own, and then joins them with the original
// error at the end.
func (s *extractState) joinErrors(baseErr error, errsIn ...error) error {
	var errsOut []error

	// Always include basic context
	errsOut = append(errsIn, fmt.Errorf("json_path=%s", s.selector))

	if s.position < len(s.segments) {
		errsOut = append(errsOut,
			fmt.Errorf("segment=%s", s.segments[s.position]),
			fmt.Errorf("segment_position=%d", s.position),
		)
	}

	if len(s.pathProgress) > 0 {
		errsOut = append(errsOut, fmt.Errorf("path_progress=%v", s.pathProgress))
	}

	// Include readable JSON context for debugging
	errsOut = append(errsOut, fmt.Errorf("condensed_json=%s", s.condensedJSON()))

	// Add the original error if provided
	if baseErr != nil {
		errsOut = append(errsOut, baseErr)
	}
	return errors.Join(errsOut...)
}
