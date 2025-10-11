// Package pathvars/template provides path template parsing and matching functionality.
// Templates represent parsed URL patterns with parameters that can be matched against
// incoming HTTP requests to extract parameter values.
package pathvars

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// ParsedTemplate represents a parsed path template with parameters and compiled
// regex. Template strings like "/users/{id:int}/posts/{slug:string}" can be
// parsed into a ParsedTemplate to allow for matching incoming request paths to
// extract parameter values.
type ParsedTemplate struct {
	// raw stores the original template string for reference and error reporting.
	raw string

	// segments contains the parsed path segments, both literal and parameter segments.
	segments []Segment

	// params maps parameter names to their definitions for validation and extraction.
	params map[Identifier]Parameter

	// paramNames contain parameter names in order they occur in the path
	paramNames []Identifier

	// regex is the compiled regular expression used for efficient path matching.
	regex *regexp.Regexp
}

// Template returns the string representation of the template that was parsed but
// as a string-derived type Template.  See String() comments for more details.
func (t *ParsedTemplate) Template() Template {
	// TODO: Verify if we should just return RAW, or if we should assemble from parsed parts.
	// If we do we can use Substitute() and pass in the template variables as if they were values.
	return Template(t.raw)
}

// String returns the string value of a parsed template which should just be what
// is raw contains. However, there may be a difference between the raw string and
// a recomposed string, so we need to be vigilant. This is currently (2025-10-05)
// only called implicitly from a fmt.Sprintf() from within apipkg.ParseEndpoint()
// which is the package pathvars was originally developed for. Note that I added
// Template() first then realized this was likely better as a fmt.Stringer method
// so added it too, but did not remove Template() even though I am not currently
// using it simply because it returns a Template type vs. a string type.
func (t *ParsedTemplate) String() string {
	return t.raw
}

// Match attempts to match a path and query string against this template.
// Returns a ValuesMap containing extracted parameter values and a boolean indicating
// whether the match was successful. Both path parameters (from URL segments) and
// query parameters are extracted and validated according to their type constraints.
func (t *ParsedTemplate) Match(path, query string) (valuesMap ValuesMap, matched bool, err error) {
	var errs []error
	var matchedPath, matchedQuery bool

	// First, match path parameters using regex
	matchedPath, err = t.matchPathParameters(path, &valuesMap)
	if err != nil {
		errs = append(errs, err)
	}
	matchedQuery, err = t.matchQueryParameters(query, &valuesMap)
	if err != nil {
		errs = append(errs, err)
	}
	if len(valuesMap) == 0 {
		valuesMap = nil
	}
	return valuesMap, matchedPath && matchedQuery, errors.Join(errs...)
}

// matchPathParameters matches path parameters using regex and adds them to vars.
// Returns false if the path doesn't match the template or if parameter validation fails.
func (t *ParsedTemplate) matchPathParameters(path string, valuesMap *ValuesMap) (matched bool, err error) {
	var matches []string
	var n int
	var name Identifier
	var value string
	var param Parameter
	var exists bool
	var errs []error

	if t.regex == nil {
		// No path regex means no path parameters
		matched = true
		goto end
	}

	matches = t.regex.FindStringSubmatch(path)
	if matches == nil {
		// No match is not an error, just no match
		matched = false
		goto end
	}
	matched = true

	// Extract parameters from regex groups
	n = 1 // Skip full match at index 0
	for _, segment := range t.segments {
		if !segment.IsParameter() {
			continue
		}

		if n >= len(matches) {
			goto end
		}

		// We currently only support one parameter per segment
		name = segment.Parameters[0].Name
		value = matches[n]

		// Validate parameter type and constraints
		param, exists = t.params[name]
		if exists && param.location == PathLocation {
			err = t.validateParameter(param, value, path, PathLocation)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(*valuesMap) == 0 {
			*valuesMap = make(ValuesMap)
		}
		(*valuesMap)[name] = value
		n++
	}
	err = errors.Join(errs...)
end:
	return matched, err
}

// matchQueryParameters matches query parameters and adds them to vars.
// Returns false if required parameters are missing or if validation fails.
// Optional parameters are handled gracefully with default values when provided.
func (t *ParsedTemplate) matchQueryParameters(query string, valuesMap *ValuesMap) (matched bool, err error) {
	var queryValues url.Values
	var param Parameter
	var name Identifier
	var value string
	var values []string
	var found bool
	var errs []error
	var addValue func(any)

	// ParseBytes query string
	if query != "" {
		queryValues, err = url.ParseQuery(query)

		if err != nil {
			err = errors.Join(ErrInvalidURLQueryString, fmt.Errorf("url_query=%s", query), err)
			goto end
		}

		if queryValues == nil {
			queryValues = make(url.Values)
		}
	}

	matched = true

	addValue = func(value any) {
		if len(*valuesMap) == 0 {
			*valuesMap = make(ValuesMap)
		}
		(*valuesMap)[name] = value
	}

	// Check each query parameter in the template
	for name, param = range t.params {
		if param.location != QueryLocation {
			continue
		}

		// Check if parameter is present in query string
		values, found = queryValues[string(name)]
		switch {
		case found && len(values) > 0:
			// Use the first value if multiple are provided
			value = values[0]

			// Validate matched parameter
			err = t.validateParameter(param, value, query, QueryLocation)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			addValue(value)

		case param.Optional:
			// Optional parameter not provided
			if param.DefaultValue != nil {
				// Use default value
				addValue(*param.DefaultValue)
			}

		default:
			// If no default value, simply omit from valuesMap (empty string behavior)
			matched = false
			// Required parameter not provided
			errs = append(errs, errors.Join(ErrRequiredParameterNotProvided,
				fmt.Errorf("parameter_name=%s", name),
				fmt.Errorf("data_type=%s", param.dataType.Slug()),
				fmt.Errorf("http_status=%d", http.StatusUnprocessableEntity),
				fmt.Errorf("parameter_location=%s", param.location),
			))
			continue
		}
	}
	err = errors.Join(errs...)
end:
	return matched, err
}

// validateParameter validates a parameter value against its type and constraints.
// Returns an error with detailed context if validation fails.
func (t *ParsedTemplate) validateParameter(param Parameter, value, source string, location LocationType) (err error) {
	var errs []error

	// Validate data type
	err = validateDataType(value, param.dataType)
	if err != nil {
		err = errors.Join(ErrInvalidDataType, err)
		goto end
	}

	// Validate constraints
	for _, constraint := range param.constraints {
		err = constraint.Validate(value)
		if err != nil {
			errs = append(errs, errors.Join(ErrConstraintValidationFailed, err))
		}
	}
	err = errors.Join(errs...)
end:
	if err != nil {
		paramType := param.dataType.WithIndefiniteArticle()

		detail := fmt.Sprintf("Parameter '%s' expected %s type but got '%s'",
			param.Name,
			paramType,
			value,
		)

		suggestion := fmt.Sprintf("Use %s for '%s' like %v, for example: %s",
			paramType,
			param.Name,
			param.dataType.Example(),
			t.Example(),
		)

		// Create RFC 9457 response with extension
		pve := ParameterValidationError{
			Err:              ErrInvalidParameterValue,
			HTTPStatus:       http.StatusUnprocessableEntity,
			Detail:           detail,
			Instance:         source,
			Parameter:        string(param.Name),
			ExpectedType:     string(param.dataType.Slug()),
			ReceivedValue:    value,
			Location:         location.Slug(),
			Suggestion:       suggestion,
			EndpointTemplate: t.raw,
		}

		err = errors.Join(ErrInvalidParameter, pve, err)
	}
	return err
}

type ParameterValidationError struct {
	Err              error
	HTTPStatus       int
	Instance         string
	Detail           string
	Parameter        string
	ExpectedType     string
	ReceivedValue    any
	Location         string
	Suggestion       string
	EndpointTemplate string
}

func (e ParameterValidationError) Error() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", e.Err.Error(),
		fmt.Sprintf("http_status=%d", e.HTTPStatus),
		fmt.Sprintf("instance=%s", e.Instance),
		fmt.Sprintf("detail=%s", e.Detail),
		fmt.Sprintf("parameter=%s", e.Parameter),
		fmt.Sprintf("expected_type=%s", e.ExpectedType),
		fmt.Sprintf("received_value=%v", e.ReceivedValue),
		fmt.Sprintf("location=%s", e.Location),
		fmt.Sprintf("suggestion=%s", e.Suggestion),
		fmt.Sprintf("endpoint_template=%s", e.EndpointTemplate),
	)
}

// Parameters returns all parameters in the template.
// TODO: Implementation needed - should return parameters in order of appearance.
func (t *ParsedTemplate) Parameters() (params []Parameter) {
	// Return parameters in order of appearance
	panic("IMPLEMENT ME")
	return params
}

// Validate checks parameter values against the template requirements.
// TODO: Implementation needed - should validate each parameter value.
func (t *ParsedTemplate) Validate(params map[Identifier]any) (err error) {
	// Validate each parameter value
	panic("IMPLEMENT ME")
	return err
}

// Substitute builds a path from parameter values by replacing template placeholders.
// TODO: Implementation needed - should build path by substituting values.
func (t *ParsedTemplate) Substitute(values map[Identifier]any) (result string, err error) {
	var errs []error
	sb := strings.Builder{}
	n := 0
	for _, seg := range t.segments {
		sb.WriteByte('/')
		if seg.IsLiteral() {
			sb.WriteString(seg.Raw)
			continue
		}
		if seg.Prefix != "" {
			sb.WriteString(seg.Prefix)
		}
		// We currently only support one parameter per segment
		paramName := seg.Parameters[0].Name
		value, ok := values[paramName]
		if !ok {
			errs = append(errs, errors.Join(ErrParameterNotFoundInValuesMap,
				fmt.Errorf("parameter_name=%s", paramName),
				fmt.Errorf("values_map=%v", values),
			))
			continue
		}
		sb.WriteString(fmt.Sprintf("%v", value))
		if seg.Suffix != "" {
			sb.WriteString(seg.Suffix)
		}
		n++
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
		goto end
	}
	result = sb.String()
end:
	return result, err
}

func (t *ParsedTemplate) Example() (result string) {
	params := make(map[Identifier]any, len(t.params))
	for name, param := range t.params {
		params[name] = param.dataType.Example()
	}
	result, _ = t.Substitute(params)
	return result
}
