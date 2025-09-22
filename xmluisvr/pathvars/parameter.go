package pathvars

import (
	"errors"
	"fmt"
	"strings"
)

type ParameterType int

const (
	UnspecifiedParameterType ParameterType = iota
	PathParameter
	QueryParameter
)

// Parameter represents a path parameter
type Parameter struct {
	name         string
	paramType    ParameterType
	dataType     PVDataType
	constraints  []Constraint
	position     int
	original     string
	multiSegment bool
	optional     bool
	defaultValue *string
}

func NewParameter(args ParameterArgs) Parameter {
	return Parameter{
		name:         args.Name,
		paramType:    args.ParamType,
		dataType:     args.DataType,
		constraints:  args.Constraints,
		position:     args.Position,
		original:     args.Original,
		multiSegment: args.MultiSegment,
		optional:     args.Optional,
		defaultValue: args.DefaultValue,
	}
}

type ParameterArgs struct {
	Name         string
	ParamType    ParameterType
	DataType     PVDataType
	Constraints  []Constraint
	Position     int
	Original     string
	MultiSegment bool
	Optional     bool
	DefaultValue *string
}

// ParseParameter parses a parameter specification like {id:int:range[1..100]} or {date*:date:yyyy/mm/dd}
// Also supports optional parameters: {name?:type} or {name?default:type:constraints}
func ParseParameter(spec string, position int) (p *Parameter, err error) {
	var content string
	var parts []string
	var name string
	var dataType PVDataType
	var paramType ParameterType
	var constraints []Constraint
	var multiSegment bool
	var optional bool
	var defaultValue *string
	var hasDoubleColon bool

	// Parse the {name:type:constraints} or {name*:type:constraints} format
	// Return Parameter object with parsed components

	// Remove braces
	if len(spec) < 2 || spec[0] != '{' || spec[len(spec)-1] != '}' {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("position=%d", position),
			fmt.Errorf("reason=%s", "parameter must be enclosed in braces"),
		)
		goto end
	}

	content = spec[1 : len(spec)-1]
	if content == "" {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("position=%d", position),
			fmt.Errorf("reason=%s", "empty parameter content"),
		)
		goto end
	}

	// Handle different parameter syntax patterns:
	// 1. {name} -> infer type from name if it matches a data type
	// 2. {name:type} -> explicit type
	// 3. {name:type:constraint} -> explicit type with constraint
	// 4. {name::constraint} -> infer type from name, constraint (double colon)
	// Split by colon, but handle special case of double colon for implicit type
	if strings.Contains(content, "::") {
		hasDoubleColon = true
		// Replace :: with :IMPLICIT: as a marker, then split normally
		content = strings.Replace(content, "::", ":IMPLICIT:", 1)
	}

	// Split by colon, but only split on first two colons to handle constraints with colons
	// e.g., "name:type:hh:mm:ss" -> ["name", "type", "hh:mm:ss"]
	parts = strings.SplitN(content, ":", 3)

	// Parse the first part which may contain name, optional marker (?), and default value
	// Possible formats:
	// - "name" -> required parameter
	// - "name?" -> optional parameter, no default
	// - "name?default" -> optional parameter with default value
	// - "name*" -> multi-segment required parameter
	// - "name*?" -> multi-segment optional parameter, no default
	// - "name*?default" -> multi-segment optional parameter with default
	name, optional, defaultValue, multiSegment, err = parseNamePart(parts[0])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("content=%q", content),
			fmt.Errorf("position=%d", position),
		)
		goto end
	}

	// Determine data type based on syntax
	dataType = StringType // default

	if len(parts) == 1 {
		// Pattern: {name} -> try to infer type from name
		if inferredType, canInfer := InferDataTypeFromName(name); canInfer {
			dataType = inferredType
		}
		// If can't infer, use default StringType
	} else if len(parts) > 1 {
		if hasDoubleColon && parts[1] == "IMPLICIT" {
			// Pattern: {name::constraint} -> infer type from name
			if inferredType, canInfer := InferDataTypeFromName(name); canInfer {
				dataType = inferredType
			} else {
				err = errors.Join(
					ErrInvalidParameter,
					fmt.Errorf("spec=%q", spec),
					fmt.Errorf("paramName=%q", name),
					fmt.Errorf("reason=%s", "cannot infer type from parameter name with :: syntax"),
				)
				goto end
			}
		} else if parts[1] != "" {
			// Pattern: {name:type} or {name:type:constraint} -> explicit type
			dataType, err = ParsePVDataType(parts[1])
			if err != nil {
				err = errors.Join(
					err,
					fmt.Errorf("spec=%q", spec),
					fmt.Errorf("paramName=%q", name),
					fmt.Errorf("typeStr=%q", parts[1]),
					fmt.Errorf("position=%d", position),
				)
				goto end
			}
		}
	}

	// Get constraints (optional)
	if len(parts) > 2 && parts[2] != "" {
		constraints, err = ParseConstraints(parts[2], dataType)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("spec=%q", spec),
				fmt.Errorf("paramName=%q", name),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("constraintSpec=%q", parts[2]),
				fmt.Errorf("position=%d", position),
			)
			goto end
		}
	}

	// Validate default value if provided
	if defaultValue != nil {
		err = validateDataType(*defaultValue, dataType)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("spec=%q", spec),
				fmt.Errorf("paramName=%q", name),
				fmt.Errorf("defaultValue=%q", *defaultValue),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "default value validation failed"),
			)
			goto end
		}

		// Also validate against constraints
		for _, constraint := range constraints {
			err = constraint.Validate(*defaultValue)
			if err != nil {
				err = errors.Join(
					err,
					fmt.Errorf("spec=%q", spec),
					fmt.Errorf("paramName=%q", name),
					fmt.Errorf("defaultValue=%q", *defaultValue),
					fmt.Errorf("constraint=%s", constraint.String()),
					fmt.Errorf("reason=%s", "default value constraint validation failed"),
				)
				goto end
			}
		}
	}

	p = &Parameter{
		name:         name,
		paramType:    paramType,
		dataType:     dataType,
		constraints:  constraints,
		position:     position,
		original:     spec,
		multiSegment: multiSegment,
		optional:     optional,
		defaultValue: defaultValue,
	}

end:
	return p, err
}

// parseNamePart parses the name part of a parameter which may contain:
// - name -> required parameter
// - name? -> optional parameter, no default
// - name?default -> optional parameter with default value
// - name* -> multi-segment required parameter
// - name*? -> multi-segment optional parameter, no default
// - name*?default -> multi-segment optional parameter with default
func parseNamePart(namePart string) (name string, optional bool, defaultValue *string, multiSegment bool, err error) {
	var questionPos int
	var starPos int
	var hasQuestion bool
	var hasStar bool
	var defaultVal string

	if namePart == "" {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("namePart=%q", namePart),
			fmt.Errorf("reason=%s", "parameter name cannot be empty"),
		)
		goto end
	}

	// Find positions of special characters
	questionPos = strings.Index(namePart, "?")
	starPos = strings.Index(namePart, "*")
	hasQuestion = questionPos != -1
	hasStar = starPos != -1

	// Validate character ordering: name comes first, then *, then ?
	if hasStar && hasQuestion && starPos > questionPos {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("namePart=%q", namePart),
			fmt.Errorf("reason=%s", "invalid syntax: '*' must come before '?' in parameter name"),
		)
		goto end
	}

	// Extract name (everything before first special character)
	if hasStar && (!hasQuestion || starPos < questionPos) {
		name = namePart[:starPos]
		multiSegment = true
	} else if hasQuestion {
		name = namePart[:questionPos]
	} else {
		name = namePart
	}

	// Validate name is not empty
	if name == "" {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("namePart=%q", namePart),
			fmt.Errorf("reason=%s", "parameter name cannot be empty before special characters"),
		)
		goto end
	}

	// Handle optional marker and default value
	if hasQuestion {
		optional = true
		// Extract default value (everything after ?)
		if hasStar && starPos < questionPos {
			// Pattern: name*?default
			defaultVal = namePart[questionPos+1:]
		} else {
			// Pattern: name?default
			defaultVal = namePart[questionPos+1:]
		}

		// If there's content after ?, it's a default value
		if defaultVal != "" {
			defaultValue = &defaultVal
		}
	}

end:
	return name, optional, defaultValue, multiSegment, err
}

// DataType returns the parameter's data type
func (p *Parameter) DataType() PVDataType {
	return p.dataType
}

// Name returns the parameter's name
func (p *Parameter) Name() string {
	return p.name
}

// IsOptional returns true if the parameter is optional
func (p *Parameter) IsOptional() bool {
	return p.optional
}

// IsMultiSegment returns true if the parameter can span multiple path segments
func (p *Parameter) IsMultiSegment() bool {
	return p.multiSegment
}

// DefaultValue returns the parameter's default value if it has one
func (p *Parameter) DefaultValue() *string {
	return p.defaultValue
}
