// Package pathvars/parser provides template parsing functionality for converting
// path template strings into Template objects with compiled regular expressions
// and parameter definitions. It handles complex parsing scenarios including
// nested braces, multi-segment parameters, and query parameter extraction.
package pathvars

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// ParseTemplate parses a template string like "/users/{id:int}/posts?{limit?10:int}"
// into a Template object with compiled regex and parameter definitions.
// Returns an error if the template syntax is invalid.
func ParseTemplate(template common.URLPath) (t *Template, err error) {
	var segments []Segment
	var params map[common.Identifier]Parameter
	var regex *regexp.Regexp

	segments, params, err = parseSegments(template)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("template=%q", template),
		)
		goto end
	}

	regex, err = buildRegex(segments, params)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("template=%q", template),
			fmt.Errorf("segments=%v", segments),
		)
		goto end
	}

	t = &Template{
		raw:      string(template),
		segments: segments,
		params:   params,
		regex:    regex,
	}

end:
	return t, err
}

// parseSegments splits a template into segments and extracts parameters.
// Handles both path and query portions of the template, parsing each
// according to their specific syntax rules.
func parseSegments(template common.URLPath) (segments []Segment, params map[common.Identifier]Parameter, err error) {
	var pathPart, queryPart string
	var pathSegments []Segment
	var pathParams, queryParams map[common.Identifier]Parameter
	var position int

	params = make(map[common.Identifier]Parameter)

	if template == "" {
		err = errors.Join(
			ErrInvalidTemplate,
			fmt.Errorf("template=%q", template),
			fmt.Errorf("reason=%s", "empty template"),
		)
		goto end
	}

	// Split template into path and query parts at the first '?' that's not inside braces
	pathPart, queryPart, err = splitPathAndQuery(template)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("template=%q", template),
		)
		goto end
	}

	// ParseBytes path segments
	pathSegments, pathParams, err = parsePathPart(pathPart)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("template=%q", template),
			fmt.Errorf("path_part=%q", pathPart),
		)
		goto end
	}

	// Set position for path parameters and add to combined params map
	position = 0
	for name, param := range pathParams {
		param.useType = PathUseType
		param.position = position
		params[name] = param
		position++
	}

	// ParseBytes query parameters if present
	if queryPart != "" {
		queryParams, err = parseQueryPart(queryPart, position)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("template=%q", template),
				fmt.Errorf("query_part=%q", queryPart),
			)
			goto end
		}

		// Add query parameters to combined params map
		for name, param := range queryParams {
			param.useType = QueryUseType
			params[name] = param
		}
	}

	segments = pathSegments

end:
	return segments, params, err
}

// buildRegex creates a regex pattern from template segments for efficient path matching.
// Handles both regular parameters and multi-segment parameters that can span
// multiple path segments.
func buildRegex(segments []Segment, params map[common.Identifier]Parameter) (re *regexp.Regexp, err error) {
	var sb strings.Builder
	var segment Segment
	var paramName common.Identifier
	var param Parameter
	var exists bool

	// Build regex string from segments
	sb.WriteByte('^')

	for _, segment = range segments {
		sb.WriteByte('/')
		if !segment.IsParameter() {
			// Literal segments - escape special regex characters
			sb.WriteString(regexp.QuoteMeta(string(segment)))
			continue
		}
		// Extract parameter name to check if it's multi-segment
		paramName = extractParamName(string(segment))
		param, exists = params[paramName]

		// Regular parameters capture any non-slash characters
		captureRegex := "([^/]+)"
		if exists && param.MultiSegment {
			// Multi-segment parameters capture non-slash chars optionally followed by more segments
			captureRegex = "([^/]+(?:/[^/]+)*)"
		}
		sb.WriteString(captureRegex)
	}
	sb.WriteByte('$')

	// Compile and return regex
	re, err = regexp.Compile(sb.String())

	return re, err
}

// parsePathSegments splits a path template into segments, being careful not to split
// on slashes that are inside parameter constraint definitions like {date:date:yyyy/mm/dd}.
// Handles nested braces and validates brace matching.
func parsePathSegments(template string) (segments []string, err error) {
	var result []string
	var currentSegment strings.Builder
	var i int
	var inBraces bool
	var braceDepth int

	// Skip leading slash if present
	if len(template) > 0 && template[0] == '/' {
		i = 1
	}

	for i < len(template) {
		char := template[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
			currentSegment.WriteByte(char)
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
				currentSegment.WriteByte(char)
			} else {
				// Unmatched closing brace - this should be an error for consistency
				err = errors.Join(
					ErrInvalidParameter,
					fmt.Errorf("template=%q", template),
					fmt.Errorf("position=%d", i),
					fmt.Errorf("char=%c", char),
					fmt.Errorf("reason=%s", "unmatched closing brace"),
				)
				goto end
			}
		case '/':
			if inBraces {
				// Inside braces, keep the slash as part of the segment
				currentSegment.WriteByte(char)
			} else {
				// Outside braces, this is a segment separator
				if currentSegment.Len() > 0 {
					result = append(result, currentSegment.String())
					currentSegment.Reset()
				}
			}
		default:
			currentSegment.WriteByte(char)
		}
		i++
	}

	// Add the final segment if there is one
	if currentSegment.Len() > 0 {
		result = append(result, currentSegment.String())
	}

	// Validate that braces are balanced - only error on unmatched opening braces
	if braceDepth > 0 {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("template=%q", template),
			fmt.Errorf("brace_depth=%d", braceDepth),
			fmt.Errorf("reason=%s", "unmatched opening brace(s)"),
		)
		goto end
	}

	segments = result

end:
	return segments, err
}

// splitPathAndQuery splits a template into path and query parts at the first '?'
// that's not inside braces. This allows query parameters to contain '?' characters
// within their constraint definitions.
func splitPathAndQuery(template common.URLPath) (pathPart, queryPart string, err error) {
	var i int
	var inBraces bool
	var braceDepth int

	for i < len(template) {
		char := template[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
			} else {
				err = errors.Join(
					ErrInvalidParameter,
					fmt.Errorf("template=%q", template),
					fmt.Errorf("position=%d", i),
					fmt.Errorf("reason=%s", "unmatched closing brace"),
				)
				goto end
			}
		case '?':
			if !inBraces {
				// Found the split point
				pathPart = string(template[:i])
				queryPart = string(template[i+1:])
				goto end
			}
		}
		i++
	}

	// Validate that braces are balanced
	if braceDepth > 0 {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("template=%q", template),
			fmt.Errorf("brace_depth=%d", braceDepth),
			fmt.Errorf("reason=%s", "unmatched opening brace(s)"),
		)
		goto end
	}

	// No query part found
	pathPart = string(template)

end:
	return pathPart, queryPart, err
}

// parsePathPart parses the path portion of a template into segments and parameters.
// Extracts parameter definitions from path segments and validates their syntax.
func parsePathPart(pathPart string) (segments []Segment, params map[common.Identifier]Parameter, err error) {
	var parts []string
	var part string
	var segment Segment
	var param Parameter
	var position int

	params = make(map[common.Identifier]Parameter)

	// ParseBytes path segments using existing logic
	parts, err = parsePathSegments(pathPart)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("path_part=%s", pathPart),
		)
		goto end
	}

	position = 0
	for _, part = range parts {
		if part == "" {
			continue // Skip empty parts (like leading slash)
		}

		segment = Segment(part)
		segments = append(segments, segment)

		// Check if this segment is a parameter
		if !segment.IsParameter() {
			continue
		}

		param, err = ParseParameter(part, PathUseType, position)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("path_part=%s", pathPart),
				fmt.Errorf("segment=%s", part),
				fmt.Errorf("position=%d", position),
			)
			goto end
		}
		params[param.Name] = param
		position++
	}

end:
	return segments, params, err
}

// parseQueryPart parses the query portion of a template like "{owner:email}&{limit?10:int}".
// Extracts parameter definitions from query parameter specifications.
func parseQueryPart(queryPart string, startPosition int) (params map[common.Identifier]Parameter, err error) {
	var queryParams []string
	var paramSpec string
	var param Parameter
	var position int

	params = make(map[common.Identifier]Parameter)

	// Split query part by '&' to get individual parameters, being careful of braces
	queryParams, err = parseQueryParameters(queryPart)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("query_part=%s", queryPart),
		)
		goto end
	}

	position = startPosition
	for _, paramSpec = range queryParams {
		if paramSpec == "" {
			continue
		}

		param, err = ParseParameter(paramSpec, QueryUseType, position)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("query_part=%s", queryPart),
				fmt.Errorf("parameter_spec=%s", paramSpec),
				fmt.Errorf("position=%d", position),
			)
			goto end
		}
		params[param.Name] = param
		position++
	}

end:
	return params, err
}

// parseQueryParameters splits query parameters by '&' while respecting braces.
// This allows constraint definitions to contain '&' characters without being
// treated as parameter separators.
func parseQueryParameters(queryPart string) (parameters []string, err error) {
	var result []string
	var currentParam strings.Builder
	var i int
	var inBraces bool
	var braceDepth int

	for i < len(queryPart) {
		char := queryPart[i]

		switch char {
		case '{':
			inBraces = true
			braceDepth++
			currentParam.WriteByte(char)
		case '}':
			if braceDepth > 0 {
				braceDepth--
				if braceDepth == 0 {
					inBraces = false
				}
				currentParam.WriteByte(char)
			} else {
				err = errors.Join(
					ErrInvalidParameter,
					fmt.Errorf("query_part=%s", queryPart),
					fmt.Errorf("position=%d", i),
					fmt.Errorf("char=%c", char),
					fmt.Errorf("reason=%s", "unmatched closing brace"),
				)
				goto end
			}
		case '&':
			if inBraces {
				// Inside braces, keep the & as part of the parameter
				currentParam.WriteByte(char)
			} else {
				// Outside braces, this is a parameter separator
				if currentParam.Len() > 0 {
					result = append(result, currentParam.String())
					currentParam.Reset()
				}
			}
		default:
			currentParam.WriteByte(char)
		}
		i++
	}

	// Add the final parameter if there is one
	if currentParam.Len() > 0 {
		result = append(result, currentParam.String())
	}

	// Validate that braces are balanced
	if braceDepth > 0 {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("query_part=%s", queryPart),
			fmt.Errorf("brace_depth=%d", braceDepth),
			fmt.Errorf("reason=%s", "unmatched opening brace(s)"),
		)
		goto end
	}

	parameters = result

end:
	return parameters, err
}
