// Package pathvars/template provides path template parsing and matching functionality.
// Templates represent parsed URL patterns with parameters that can be matched against
// incoming HTTP requests to extract parameter values.
package pathvars

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// Template represents a parsed path template with parameters and compiled regex.
// Templates are created from path strings like "/users/{id:int}/posts/{slug:string}"
// and can match incoming request paths to extract parameter values.
type Template struct {
	// raw stores the original template string for reference and error reporting.
	raw string

	// segments contains the parsed path segments, both literal and parameter segments.
	segments []Segment

	// params maps parameter names to their definitions for validation and extraction.
	params map[common.Identifier]Parameter

	// regex is the compiled regular expression used for efficient path matching.
	regex *regexp.Regexp
}

// Match attempts to match a path and query string against this template.
// Returns a VarsMap containing extracted parameter values and a boolean indicating
// whether the match was successful. Both path parameters (from URL segments) and
// query parameters are extracted and validated according to their type constraints.
func (t *Template) Match(path, queryString string) (vars VarsMap, matched bool) {
	vars = make(VarsMap)
	matched = false

	// First, match path parameters using regex
	if !t.matchPathParameters(path, vars) {
		goto end
	}

	// Then, match query parameters
	if !t.matchQueryParameters(queryString, vars) {
		goto end
	}

	matched = true

end:
	return vars, matched
}

// matchPathParameters matches path parameters using regex and adds them to vars.
// Returns false if the path doesn't match the template or if parameter validation fails.
func (t *Template) matchPathParameters(path string, vars VarsMap) bool {
	var matches []string
	var i int
	var name common.Identifier
	var value string
	var param Parameter
	var exists bool
	var err error

	if t.regex == nil {
		return true // No path regex means no path parameters
	}

	matches = t.regex.FindStringSubmatch(path)
	if matches == nil {
		return false // Path doesn't match
	}

	// Extract parameters from regex groups
	i = 1 // Skip full match at index 0
	for _, segment := range t.segments {
		if !segment.IsParameter() {
			continue
		}

		if i >= len(matches) {
			return false
		}

		name = extractParamName(string(segment))
		value = matches[i]

		// Validate parameter type and constraints
		param, exists = t.params[name]
		if exists && param.useType == PathUseType {
			err = t.validateParameter(param, value, path)
			if err != nil {
				return false
			}
		}

		vars[name] = value
		i++
	}

	return true
}

// matchQueryParameters matches query parameters and adds them to vars.
// Returns false if required parameters are missing or if validation fails.
// Optional parameters are handled gracefully with default values when provided.
func (t *Template) matchQueryParameters(queryString string, vars VarsMap) bool {
	var queryValues url.Values
	var param Parameter
	var name common.Identifier
	var value string
	var values []string
	var found bool
	var err error

	// ParseBytes query string
	if queryString != "" {
		queryValues, err = url.ParseQuery(queryString)
		if err != nil {
			return false
		}
	} else {
		queryValues = make(url.Values)
	}

	// Check each query parameter in the template
	for name, param = range t.params {
		if param.useType != QueryUseType {
			continue
		}

		// Check if parameter is present in query string
		values, found = queryValues[string(name)]
		if found && len(values) > 0 {
			// Use the first value if multiple are provided
			value = values[0]

			// Validate parameter
			err = t.validateParameter(param, value, queryString)
			if err != nil {
				return false
			}

			vars[name] = value
		} else if param.Optional {
			// Optional parameter not provided
			if param.DefaultValue != nil {
				// Use default value
				vars[name] = *param.DefaultValue
			}
			// If no default value, simply omit from vars (empty string behavior)
		} else {
			// Required parameter not provided
			return false
		}
	}

	return true
}

// validateParameter validates a parameter value against its type and constraints.
// Returns an error with detailed context if validation fails.
func (t *Template) validateParameter(param Parameter, value, context string) error {
	var err error

	// Validate data type
	err = validateDataType(value, param.dataType)
	if err != nil {
		return errors.Join(
			err,
			fmt.Errorf("context=%q", context),
			fmt.Errorf("template=%q", t.raw),
			fmt.Errorf("parameter=%q", param.Name),
			fmt.Errorf("value=%q", value),
			fmt.Errorf("data_type=%v", param.dataType),
			fmt.Errorf("reason=%s", "data type validation failed"),
		)
	}

	// Validate constraints
	for _, constraint := range param.constraints {
		err = constraint.Validate(value)
		if err != nil {
			return errors.Join(
				err,
				fmt.Errorf("context=%q", context),
				fmt.Errorf("template=%q", t.raw),
				fmt.Errorf("parameter=%q", param.Name),
				fmt.Errorf("value=%q", value),
				fmt.Errorf("constraint=%s", constraint.String()),
				fmt.Errorf("reason=%s", "constraint validation failed"),
			)
		}
	}

	return nil
}

// Parameters returns all parameters in the template.
// TODO: Implementation needed - should return parameters in order of appearance.
func (t *Template) Parameters() (params []Parameter) {
	// Return parameters in order of appearance
	return params
}

// Validate checks parameter values against the template requirements.
// TODO: Implementation needed - should validate each parameter value.
func (t *Template) Validate(params map[string]string) (err error) {
	// Validate each parameter value
	return err
}

// Substitute builds a path from parameter values by replacing template placeholders.
// TODO: Implementation needed - should build path by substituting values.
func (t *Template) Substitute(values map[string]string) (result string, err error) {
	// Build path by substituting values
	return result, err
}

// extractParamName extracts the parameter name from a segment like {id:int} or {date*:date:format}.
// Returns just the parameter name without type specifications or multi-segment markers.
func extractParamName(segment string) (name common.Identifier) {
	var content string
	var idx int

	// Remove braces and extract name (before first colon if any)
	if len(segment) < 2 {
		goto end
	}

	if segment[0] != '{' {
		goto end
	}

	if segment[len(segment)-1] != '}' {
		goto end
	}

	content = segment[1 : len(segment)-1]

	// Get just the name part (before first colon)
	idx = strings.Index(content, ":")
	if idx == -1 {
		idx = len(content)
	}
	name = common.Identifier(content[:idx])
	if name == "" {
		goto end
	}

	// Remove multi-segment suffix if present
	if strings.HasSuffix(string(name), "*") {
		name = name[:len(name)-1]
	}

end:
	return name
}
