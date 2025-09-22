package cfgldr

import (
	"bytes"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type APIParamsMapKey string

var _ APIParam = (*APIParamsMapValue)(nil)

type APIParamsMapValue string

func (APIParamsMapValue) APIParam() {}

var _ APIParamsMapper = (*APIParamsMap)(nil)

type APIParamsMap struct {
	OrderedMap[APIParamsMapKey, APIParamsMapValue]
}

func (pm *APIParamsMap) APIParamsMap() *APIParamsMap {
	return pm
}

func (pm *APIParamsMap) APIParamsV1() (params APIParamsV1) {
	params = APIParamsV1{}
	for name, spec := range pm.Iterator() {
		// Skip comment keys (those starting with @)
		if len(name) > 0 && name[0] == '@' {
			continue
		}
		typ, constraints, found := strings.Cut(string(spec), ":")
		if !found {
			params = append(params,
				NewAPIParamV1(string(name), string(spec)),
			)
			continue
		}
		params = append(params,
			NewAPIParamV1WithConstraints(string(name), typ, constraints),
		)
	}
	return params
}

// MarshalJSONTo encodes keys in their stored order.
func (pm *APIParamsMap) MarshalJSONTo(enc *jsontext.Encoder) (err error) {
	if pm == nil {
		// Match stdlib behavior for nil maps: {} not null.
		err = enc.WriteToken(jsontext.BeginObject)
		if err != nil {
			goto end
		}
		err = enc.WriteToken(jsontext.EndObject)
		goto end
	}
	err = enc.WriteToken(jsontext.BeginObject)
	if err != nil {
		goto end
	}
	for _, k := range pm.keys {
		// Write the name then the value using json v2’s semantic encode.
		err = jsonv2.MarshalEncode(enc, &k)
		if err != nil {
			goto end
		}
		v := pm.store[k]
		err = jsonv2.MarshalEncode(enc, &v)
		if err != nil {
			goto end
		}
	}
	err = enc.WriteToken(jsontext.EndObject)
end:
	return err
}

// UnmarshalJSON enforces all rules and detects trailing bytes inside this value.
// It also supports top-level null => empty map, and comment keys that may be
// string or array-of-strings.
func (pm *APIParamsMap) UnmarshalJSON(b []byte) (err error) {
	var dec *jsontext.Decoder
	var offset int

	// Treat top-level null as empty map; round-trip will be "{}".
	if isJSONNull(b) {
		pm.Clear()
		goto end
	}

	dec = jsontext.NewDecoder(bytes.NewReader(b))

	if dec.PeekKind() != '{' {
		err = errors.Join(
			ErrAPIParamsMapExpectedObject,
			fmt.Errorf("token_kind=%q", dec.PeekKind()),
		)
		goto end
	}

	_, err = dec.ReadToken()
	if err != nil {
		goto end
	}

	pm.Clear()

	for dec.PeekKind() != '}' {
		// 1) name
		var name string
		var valOffset int64
		var isComment bool

		err = jsonv2.UnmarshalDecode(dec, &name)
		if err != nil {
			goto end
		}
		// 2) value kind checks + decode
		valOffset = dec.InputOffset()
		isComment = len(name) > 0 && name[0] == '@'
		k := dec.PeekKind()
		switch k {
		case '{':
			// Nested object is forbidden for real params; for comments we also forbid objects.
			err = errors.Join(
				ErrAPIParamsMapCannotBeNested,
				fmt.Errorf("key=%q", name),
				fmt.Errorf("value=%s", previewNextValue(name, dec)),
				fmt.Errorf("offset=%d", valOffset),
			)
			goto end

		case '[':
			var ss []string
			if !isComment {
				err = errors.Join(
					ErrAPIParamsMapCannotContainArray,
					fmt.Errorf("key=%q", name),
					fmt.Errorf("value=%s", previewNextValue(name, dec)),
					fmt.Errorf("offset=%d", valOffset),
				)
				goto end
			}
			// For @comments, allow array of strings.
			ss, err = readStringArray(name, dec)
			if err != nil {
				err = errors.Join(
					ErrAPIParamsMapCannotContainArray,
					fmt.Errorf("key=%q", name),
					fmt.Errorf("value=%s", previewBytesArray(ss)),
					fmt.Errorf("offset=%d", valOffset),
					err,
				)
				goto end
			}
			// Store joined with newlines (simple, human-friendly). If you prefer another
			// representation, switch here. This keeps the “values are strings” invariant.
			joined := joinWithNewlines(ss)
			// Duplicate check
			if prev, dup := pm.Get(APIParamsMapKey(name)); dup {
				err = errors.Join(
					ErrAPIParamsMapDuplicateKey,
					fmt.Errorf("key=%q", name),
					fmt.Errorf("new_value=%q", joined),
					fmt.Errorf("prev_value=%q", prev),
					fmt.Errorf("offset=%d", valOffset),
				)
				goto end
			}
			pm.Set(APIParamsMapKey(name), APIParamsMapValue(joined))
			continue

		case '"':
			// String — OK for both params and comments.
			var s string
			err = jsonv2.UnmarshalDecode(dec, &s)
			if err != nil {
				goto end
			}
			prev, dup := pm.Get(APIParamsMapKey(name))
			if dup {
				err = errors.Join(
					ErrAPIParamsMapDuplicateKey,
					fmt.Errorf("key=%q", name),
					fmt.Errorf("new_value=%q", s),
					fmt.Errorf("prev_value=%q", prev),
					fmt.Errorf("offset=%d", valOffset),
				)
				goto end
			}
			pm.Set(APIParamsMapKey(name), APIParamsMapValue(s))
		case 'n', 't', 'f': // null / true / false
			// Decode to capture a printable preview for the error message.
			raw, _ := readRawValue(name, dec)
			err = errors.Join(
				ErrAPIParamsMapStringsOnly,
				fmt.Errorf("key=%q", name),
				fmt.Errorf("value=%s", string(raw)),
				fmt.Errorf("offset=%d", valOffset),
			)
			goto end
		default:
			// numbers or any other invalid token
			raw, _ := readRawValue(name, dec)
			err = errors.Join(
				ErrAPIParamsMapStringsOnly,
				fmt.Errorf("key=%q", name),
				fmt.Errorf("value=%s", string(raw)),
				fmt.Errorf("offset=%d", valOffset),
			)
			goto end
		}
	}

	// Consume '}' and then ensure no trailing non-WS bytes remain.
	_, err = dec.ReadToken()
	if err != nil {
		goto end
	}
	// dec.InputOffset() points just after '}'. If any non-space remains in b,
	// that's trailing data for THIS value; report a preview.
	offset = int(dec.InputOffset())
	if hasNonSpace(b[offset:]) {
		tr := trimLeadingSpace(b[offset:])
		err = errors.Join(
			ErrAPIParamsMapTrailingData,
			fmt.Errorf("offset=%d", offset+leadingSpaceCount(b[offset:])),
			fmt.Errorf("trailing=%q", previewBytes(tr, 120)),
		)
		goto end
	}
end:
	return err
}

// FYI: We are not this because it does not provide the level of control we want,
// but I am temporarily keeping it here for reference for other use-cases. It can
// be deleted at any time.
//// UnmarshalJSONFrom implements the streaming form. It enforces the same
//// rules as above, but (by design) does NOT attempt to detect trailing data,
//// because that would consume the parent stream. Top-level null => empty map.
//func (pm *APIParamsMap) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
//	switch dec.PeekKind() {
//	case 'n': // null -> empty
//		// consume the null
//		var x any
//		if err := jsonv2.UnmarshalDecode(dec, &x); err != nil {
//			return err
//		}
//		pm.Clear()
//		return nil
//	case '{':
//		// proceed
//	default:
//		return errors.Join(
//			ErrAPIParamsMapExpectedObject,
//			fmt.Errorf("got kind %q", dec.PeekKind()),
//		)
//	}
//
//	if _, err := dec.ReadToken(); err != nil { // '{'
//		return err
//	}
//
//	pm.Clear()
//
//	for dec.PeekKind() != '}' {
//		var name string
//		if err := jsonv2.UnmarshalDecode(dec, &name); err != nil {
//			return err
//		}
//		valOffset := dec.InputOffset()
//		isComment := len(name) > 0 && name[0] == '@'
//
//		switch k := dec.PeekKind(); k {
//		case '{':
//			return errors.Join(
//				ErrAPIParamsMapCannotBeNested,
//				fmt.Errorf("key=%q value=%s offset=%d", name, previewNextValue(name, dec), valOffset),
//			)
//		case '[':
//			if !isComment {
//				return errors.Join(
//					ErrAPIParamsMapCannotContainArray,
//					fmt.Errorf("key=%q value=%s offset=%d", name, previewNextValue(name, dec), valOffset),
//				)
//			}
//			// comment array of strings
//			ss, err := readStringArray(name, dec)
//			if err != nil {
//				return err
//			}
//			joined := joinWithNewlines(ss)
//			if prev, dup := pm.Get(APIParamsMapKey(name)); dup {
//				return errors.Join(
//					ErrAPIParamsMapDuplicateKey,
//					fmt.Errorf("key=%q new_value=%q prev_value=%q offset=%d", name, joined, prev, valOffset),
//				)
//			}
//			pm.Set(APIParamsMapKey(name), APIParamsMapValue(joined))
//
//		case '"':
//			var s string
//			if err := jsonv2.UnmarshalDecode(dec, &s); err != nil {
//				return err
//			}
//			if prev, dup := pm.Get(APIParamsMapKey(name)); dup {
//				return errors.Join(
//					ErrAPIParamsMapDuplicateKey,
//					fmt.Errorf("key=%q new_value=%q prev_value=%q offset=%d", name, s, prev, valOffset),
//				)
//			}
//			pm.Set(APIParamsMapKey(name), APIParamsMapValue(s))
//
//		case 'n', 't', 'f':
//			raw, _ := readRawValue(name, dec)
//			return errors.Join(
//				ErrAPIParamsMapStringsOnly,
//				fmt.Errorf("key=%q value=%s offset=%d", name, string(raw), valOffset),
//			)
//		default:
//			raw, _ := readRawValue(name, dec)
//			return errors.Join(
//				ErrAPIParamsMapStringsOnly,
//				fmt.Errorf("key=%q value=%s offset=%d", name, string(raw), valOffset),
//			)
//		}
//	}
//
//	_, err := dec.ReadToken() // '}'
//	return err
//}

// previewNextValue peeks the next JSON value and returns a short raw snippet.
// It DOES NOT advance the decoder permanently.
func previewNextValue(name string, dec *jsontext.Decoder) string {
	// Save offset; try to read a raw value; then we must reset the decoder.
	raw, _ := readRawValue(name, dec)
	// Reset by recreating decoder at original offset: jsontext.Decoder has no direct Seek,
	// so we only use this in error paths where exact reset isn't required afterwards.
	// (We only call preview just before returning an error.)
	return previewBytes(raw, 120)
}

// previewBytesArray returns a compact JSON-like preview of a slice of strings.
// It quotes each element, joins with commas, and truncates after 3 items,
// appending "…" if more remain. Intended for error messages where including
// the full array would be too verbose.
func previewBytesArray(ss []string) (s string) {
	var b strings.Builder
	b.WriteByte('[')
	for i, s := range ss {
		if i > 0 {
			b.WriteByte(',')
		}
		if i >= 3 {
			b.WriteString("…")
			break
		}
		b.WriteByte('"')
		b.WriteString(s)
		b.WriteByte('"')
	}
	b.WriteByte(']')
	return b.String()
}

// readRawValue reads the next JSON value from the decoder into a raw []byte
// without attempting to interpret its contents. It attaches the parameter name
// to the error (if any) for better diagnostics.
func readRawValue(name string, dec *jsontext.Decoder) (_ []byte, err error) {
	var rm jsontext.Value
	err = jsonv2.UnmarshalDecode(dec, &rm)
	if err != nil {
		err = errors.Join(err,
			fmt.Errorf("varname=%s", name),
		)
		goto end
	}
end:
	return rm, err
}

// readStringArray reads a JSON array of strings from the decoder. It enforces
// that every element is a JSON string. If a non-string element is encountered,
// the function consumes it, reports an error with context, and returns.
func readStringArray(name string, dec *jsontext.Decoder) (arr []string, err error) {

	// We know PeekKind() == '[' here. Consume opening bracket.
	_, err = dec.ReadToken()
	if err != nil { // '['
		goto end
	}
	for {
		k := dec.PeekKind()
		switch k {
		case ']':
			// End of array
			_, err = dec.ReadToken()
			goto end
		case '"':
			// Valid string element
			var s string
			err = jsonv2.UnmarshalDecode(dec, &s)
			if err != nil {
				goto end
			}
			arr = append(arr, s)
		default:
			// Invalid element: consume and report error with offending raw value.
			raw, _ := readRawValue(name, dec)
			err = errors.Join(ErrAPIParamsMapCannotContainArray,
				fmt.Errorf("comment_name=%s", name),
				fmt.Errorf("comment_value=%s", string(raw)),
			)
			goto end
		}
	}
end:
	return arr, err
}

// joinWithNewlines concatenates a slice of strings with '\n' between each.
// It pre-computes capacity for efficiency. Used to flatten comment arrays into
// a single string value while preserving multi-line formatting.
func joinWithNewlines(ss []string) (s string) {
	var line int
	var b []byte

	// Preserve each line; minimal processing (no trimming).
	// You can switch to another representation at any time.
	if len(ss) == 0 {
		goto end
	}
	line = 0
	for _, s := range ss {
		line += len(s)
	}
	line += len(ss) - 1
	b = make([]byte, 0, line)
	for i, s := range ss {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, s...)
	}
	s = string(b)
end:
	return s
}

// isPrintableASCII returns true if the string contains only ASCII printable
// runes plus newline, carriage return, or tab. Useful for validating that
// comments don’t contain control characters.
func isPrintableASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || (!unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t') {
			return false
		}
	}
	return true
}

// isJSONNull checks if the byte slice represents a JSON null literal (possibly
// surrounded by whitespace). Returns true only if it is exactly "null".
func isJSONNull(b []byte) (isNull bool) {
	i := 0
	for i < len(b) && isSpace(b[i]) {
		i++
	}
	if len(b)-i < 4 {
		goto end
	}
	if b[i] != 'n' {
		// Does not start with 'n' so cannot be 'null'
		goto end
	}
	if string(b[i:i+4]) != "null" {
		goto end
	}
	for j := i + 4; j < len(b); j++ {
		if isSpace(b[j]) {
			continue
		}
		goto end
	}
	isNull = true
end:
	return isNull
}

// isSpace reports whether a byte is one of the JSON whitespace characters
// (space, tab, newline, carriage return).
func isSpace(c byte) (isSpace bool) {
	switch c {
	case ' ', '\t', '\n', '\r':
		isSpace = true
	}
	return isSpace
}

// hasNonSpace reports whether the byte slice contains any non-whitespace byte.
func hasNonSpace(b []byte) (has bool) {
	for _, c := range b {
		if isSpace(c) {
			continue
		}
		has = true
		goto end
	}
end:
	return has
}

// trimLeadingSpace removes leading JSON whitespace characters from a slice.
func trimLeadingSpace(b []byte) []byte {
	return b[leadingSpaceCount(b):]
}

// leadingSpaceCount counts how many leading JSON whitespace characters are
// present at the start of the slice.
func leadingSpaceCount(b []byte) int {
	i := 0
	for i < len(b) && isSpace(b[i]) {
		i++
	}
	return i
}

// previewBytes returns a string preview of up to n bytes from the slice.
// If the slice is longer than n, an ellipsis is appended.
func previewBytes(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
