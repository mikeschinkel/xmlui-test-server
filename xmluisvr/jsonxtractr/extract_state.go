package jsonxtractr

import (
	"encoding/json/jsontext"
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
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONIndexOutOfRange,
			"target_index", targetIdx,
		)
		goto end
	}

	if kind != '[' {
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONPathExpectedArrayAtSegment,
			"expected_type", "array",
			"actual_type", kind.String(),
		)
		goto end
	}

	// Read array start token '['
	_, err = s.decoder.ReadToken()
	if err != nil {
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONTokenReadFailed,
			"expected_token", "array_start",
			err,
		)
		goto end
	}

	// Skip elements until we reach the target index
	currentIdx = 0
	for currentIdx < targetIdx {
		if s.decoder.PeekKind() == ']' {
			err = s.enrichError(
				ErrJSONPathTraversalFailed,
				ErrJSONIndexOutOfRange,
				"target_index", targetIdx,
				"array_length", currentIdx,
			)
			goto end
		}
		err = s.decoder.SkipValue()
		if err != nil {
			err = s.enrichError(
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				"skip_index", currentIdx,
				err,
			)
			goto end
		}
		currentIdx++
	}

	// Check if we're at the end of array before target index
	if s.decoder.PeekKind() == ']' {
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONIndexOutOfRange,
			"target_index", targetIdx,
			"array_length", currentIdx,
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
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONPathExpectedObjectAtSegment,
			"expected_type", "object",
			"actual_type", kind.String(),
		)
		goto end
	}

	// Read object start token '{'
	_, err = s.decoder.ReadToken()
	if err != nil {
		err = s.enrichError(
			ErrJSONPathTraversalFailed,
			ErrJSONTokenReadFailed,
			"expected_token", "object_start",
			err,
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
			err = s.enrichError(
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				"reading", "object_key",
				err,
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
			err = s.enrichError(
				ErrJSONPathTraversalFailed,
				ErrJSONTokenReadFailed,
				"skipping_key", key,
				err,
			)
			goto end
		}
	}

	// Key not found
	err = s.enrichError(
		ErrJSONPathTraversalFailed,
		ErrJSONPathSegmentNotFound,
		"missing_key", targetKey,
		"available_keys", availableKeys,
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

// enrichError takes sentinel errors and/or key-value pairs, adds state-specific
// context metadata, and optionally joins with a trailing cause error.
// Usage patterns:
//   - s.enrichError(ErrSentinel1, ErrSentinel2, "key", value)
//   - s.enrichError(ErrSentinel, "key", value, causeErr)
//   - s.enrichError(nil, ErrSentinel1, ErrSentinel2, "key", value)
func (s *extractState) enrichError(parts ...any) error {
	// Build a parts list: sentinels, then state context KVs, then remaining parts
	var allParts []any

	// Separate sentinels at the beginning from the rest
	sentinelCount := 0
	for i, part := range parts {
		if _, ok := part.(error); ok && i == sentinelCount {
			sentinelCount++
		} else {
			break
		}
	}

	// Start with the sentinels
	allParts = append(allParts, parts[:sentinelCount]...)

	// Add state-specific context metadata
	allParts = append(allParts,
		"json_path", s.selector,
	)

	if s.position < len(s.segments) {
		allParts = append(allParts,
			"segment", s.segments[s.position],
			"segment_position", s.position,
		)
	}

	if len(s.pathProgress) > 0 {
		allParts = append(allParts, "path_progress", s.pathProgress)
	}

	// Include readable JSON context for debugging
	allParts = append(allParts, "condensed_json", s.condensedJSON())

	// Append remaining parts (KV pairs and optional trailing cause error)
	allParts = append(allParts, parts[sentinelCount:]...)

	return NewErr(allParts...)
}
