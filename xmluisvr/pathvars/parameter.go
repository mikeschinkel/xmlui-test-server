// Package pathvars/parameter defines parameter types and parsing functionality
// for path and query parameters. Parameters can have data types, constraints,
// default values, and can be either required or optional.
package pathvars

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// ParamUseType indicates how a parameter is used in the template.
type ParamUseType int

// Parameter usage types.
const (
	// UnspecifiedParamUseType indicates the parameter usage type is not specified.
	UnspecifiedParamUseType ParamUseType = iota

	// PathUseType indicates the parameter is extracted from the URL path.
	PathUseType

	// QueryUseType indicates the parameter is extracted from the query string.
	QueryUseType

	IrrelevantParamUseType
)

// Parameter represents a path or query parameter with its type, constraints, and configuration.
// Parameters can be required or optional, have default values, and span multiple path segments.
type Parameter struct {
	// useType indicates whether this is a path or query parameter.
	useType ParamUseType

	// dataType specifies the expected data type for validation.
	dataType PVDataType

	// constraints contains validation rules applied to parameter values.
	constraints []Constraint

	// position indicates the parameter's position among path parameters for regex capture groups.
	position int

	// original stores the original parameter specification string for reference.
	original string

	nameProps
}
type nameProps = NameSpecProps

// NewParameter creates a new Parameter instance with the specified configuration.
func NewParameter(args ParameterArgs) Parameter {
	return Parameter{
		useType:     args.UseType,
		dataType:    args.DataType,
		constraints: args.Constraints,
		position:    args.Position,
		original:    args.Original,
		nameProps:   args.NameProps,
	}
}

// ParameterArgs contains arguments for creating a Parameter instance.
// This struct allows for easy parameter construction with named fields.
type ParameterArgs struct {
	// NameProps contains properties defined in the name
	NameProps NameSpecProps

	// UseType indicates if this is a path or query parameter.
	UseType ParamUseType

	// DataType specifies the expected data type.
	DataType PVDataType

	// Constraints contains validation rules.
	Constraints []Constraint

	// Position indicates the parameter's position for regex matching.
	Position int

	// Original stores the original parameter specification.
	Original string
}

func isBraceEnclosed(s string) (enclosed bool) {
	switch {
	case len(s) < 2:
		goto end
	case s[0] != '{':
		goto end
	case s[len(s)-1] != '}':
		goto end
	default:
		enclosed = true
	}
end:
	return enclosed
}

var ErrNotBraceEnclosed = errors.New("not brace enclosed with '{' and '}'")

func ParseBraceEnclosed(s string) (_ string, err error) {
	// Remove braces
	if !isBraceEnclosed(s) {
		err = errors.Join(ErrNotBraceEnclosed, fmt.Errorf("value=%s", s))
		goto end
	}

	s = s[1 : len(s)-1]
	if s == "" {
		err = errors.Join(ErrValueCannotBeEmpty, fmt.Errorf("value=%s", s))
		goto end
	}
end:
	return s, err
}

// ParseParameter parses a parameter specification like {id:int:range[1..100]} or {date*:date:yyyy/mm/dd}.
// Also supports optional parameters: {name?:type} or {name?default:type:constraints}.
// The position parameter indicates the parameter's position for regex capture group ordering.
func ParseParameter(spec string, useType ParamUseType, position int) (p Parameter, err error) {
	var content string
	var parts []string
	var dataType PVDataType
	var name string
	var constraints []Constraint
	var props *NameSpecProps

	// ParseBytes the {name:type:constraints} or {name*:type:constraints} format
	// Return Parameter object with parsed components

	if useType == UnspecifiedParamUseType {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("parameter_spec=%q", spec),
			fmt.Errorf("reason=%s", "parameter use type not specified"),
		)
		goto end
	}

	content, err = ParseBraceEnclosed(spec)
	if err != nil {
		err = errors.Join(
			ErrInvalidParameter,
			fmt.Errorf("parameter_spec=%s", spec),
			fmt.Errorf("position=%d", position),
		)
		goto end
	}

	// Split by colon, but only split on first two colons to handle constraints with colons
	// e.g., "name:type:hh:mm:ss" -> ["name", "type", "hh:mm:ss"]
	parts = strings.SplitN(content, ":", 3)
	name = parts[0]

	// ParseBytes the first part which may contain name, optional marker (?), and default value
	// Possible formats:
	// - "name" -> required parameter
	// - "name?" -> optional parameter, no default
	// - "name?default" -> optional parameter with default value
	// - "name*" -> multi-segment required parameter
	// - "name*?" -> multi-segment optional parameter, no default
	// - "name*?default" -> multi-segment optional parameter with default
	props, err = ParseNameSpecProps(parts[0])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("parameter_spec=%q", spec),
			fmt.Errorf("content=%q", content),
			fmt.Errorf("position=%d", position),
		)
		goto end
	}
	if props == nil {
		// Added this here because Goland flags props.DataType as possibly being null. I
		// don't see how it could be possible, but maybe Goland knows something I don't?
		panic(fmt.Sprintf("NameSpecProps are nil when err is also nil; spec=%s", spec))
	}
	switch {
	case props.DataType != nil:
		// Pattern: {name} -> name matched a data type
		dataType = *props.DataType
	case len(parts) == 1:
		// Pattern: {name} -> name not a data type
		dataType = DefaultPVDataType
	case len(parts) > 1:
		dataType, err = ParseParameterDataType(name, parts[1])
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("parameter_spec=%s", spec),
				fmt.Errorf("position=%d", position),
			)
			goto end
		}
	}

	// Get constraints (optional)
	if len(parts) > 2 && parts[2] != "" {
		constraints, err = ParseConstraints(parts[2], dataType)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("parameter_spec=%q", spec),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("constraint_spec=%q", parts[2]),
				fmt.Errorf("position=%d", position),
			)
			goto end
		}
	}

	// Validate default value if provided
	if props.DefaultValue != nil {
		err = validateDataType(*props.DefaultValue, dataType)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("parameter_spec=%q", spec),
				fmt.Errorf("parameter_name=%q", name),
				fmt.Errorf("default_value=%q", *props.DefaultValue),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "default value validation failed"),
			)
			goto end
		}

		// Also validate against constraints
		for _, constraint := range constraints {
			err = constraint.Validate(*props.DefaultValue)
			if err != nil {
				err = errors.Join(
					err,
					fmt.Errorf("parameter_spec=%q", spec),
					fmt.Errorf("parameter_name=%q", name),
					fmt.Errorf("default_value=%q", *props.DefaultValue),
					fmt.Errorf("data_type=%v", dataType),
					fmt.Errorf("constraint=%s", constraint.String()),
					fmt.Errorf("reason=%s", "default value constraint validation failed"),
				)
				goto end
			}
		}
	}
	p = Parameter{
		nameProps:   *props,
		useType:     useType,
		dataType:    dataType,
		constraints: constraints,
		position:    position,
		original:    spec,
	}

end:
	return p, err
}

type NameSpecProps struct {
	// Name is the parameter name used in the template and for value extraction.
	Name common.Identifier

	// MultiSegment indicates if this parameter can span multiple path segments.
	MultiSegment bool

	// Optional indicates if this parameter is optional (may be omitted).
	Optional bool

	// DefaultValue contains the default value for optional parameters.
	DefaultValue *string

	RawValue string

	DataType *PVDataType
}

func (p NameSpecProps) String() string {
	sb := strings.Builder{}
	sb.WriteString(string(p.Name))
	if p.MultiSegment {
		sb.WriteString("*")
	}
	if !p.Optional {
		goto end
	}
	sb.WriteString("?")
	if p.DefaultValue != nil {
		sb.WriteString(*p.DefaultValue)
	}
end:
	return sb.String()
}

func ParseParameterDataType(name, typ string) (dt PVDataType, err error) {
	// Determine data type based on syntax
	switch {
	case typ != "":
		// Pattern: {name:type} or  {name:type:} or {name:type:constraint} -> explicit type
		dt, err = ParsePVDataType(typ)
		if err != nil {
			err = errors.Join(
				err,
				fmt.Errorf("parameter_name=%s", name),
				fmt.Errorf("data_type=%s", typ),
			)
			goto end
		}
	default:
		// Pattern: {name:} or {name::} or {name::constraint} -> infer type from name
		inferredType := InferDataTypeFromName(name)
		if inferredType != UnspecifiedDataType {
			dt = inferredType
			goto end
		}
		dt = DefaultPVDataType
	}
end:
	return dt, err
}

// DataType returns the parameter's data type.
func (p Parameter) DataType() PVDataType {
	return p.dataType
}
