package apipkg

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

// ParseEndpointParams converts configuration parameters into EndpointParam structs.
// It handles type conversion and validation for each parameter definition.
func ParseEndpointParams(cfgParams cfgldr.APIParamsMapper, epPath pathvars.Template) (epParams []EndpointParam, err error) {
	var errs []error
	var pathVars []pathvars.ParamVar
	var pathValuesMap map[pathvars.Identifier]pathvars.ParamVar
	var paramsMap map[string]cfgldr.APIParamV1

	apiParams, ok := cfgParams.(cfgldr.APIParamsV1)
	if !ok {
		err = errors.Join(ErrInvalidAPIEndpointParameters, ErrCannotTypeAssert,
			fmt.Errorf("from_type=%T", cfgParams),
			fmt.Errorf("to_type=%T", (cfgldr.APIParamsV1)(nil)),
			fmt.Errorf("parameter_value=%v", cfgParams),
		)
		goto end
	}
	// Add any parameters that are defined ih the URL path but not lists in the array
	// of params.
	pathVars, err = pathvars.ParseParamsInTemplate(epPath)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("url_template=%s", epPath),
		)
		goto end
	}

	// Loop through all the APIParamsV1 and see if there are any vars from the Path string
	// that need to be have their Location or Constraints updated. Also check to make sure that
	// there are not conflicting types nor conflicting constraints
	pathValuesMap = pathvars.ParamVars(pathVars).Map()
	for i, p := range apiParams {
		name, err := pathvars.ParseLeadingIdentifier(p.NameSpec)
		if err != nil {
			// TODO Add regular error handling
			panic("Invalid identifier")
		}
		pv, ok := pathValuesMap[name]
		if !ok {
			apiParams[i].Location = string(apiresp.QueryLocation)
			continue
		}
		// Path var use-type is authoritative so assign the use-type from the path var to
		// the APIParamV1.
		apiParams[i].Location = string(pv.Location)
		dt, err := pathvars.ParsePVDataType(p.Type)
		if err != nil {
			// TODO Add regular error handling
			panic("Invalid type")
		}
		if dt != pv.Type {
			// TODO Add regular error handling
			panic("Type mismatch")
		}
		switch {
		case len(pv.Constraints) != 0 && p.Constraints == "":
			// If path var has constraints and APIParamV1 had no, transfer to APIParamV1
			apiParams[i].Constraints = pathvars.Constraints(pv.Constraints).String()
		case len(pv.Constraints) != 0 && p.Constraints != "":
			// If both path var has constraints and APIParamV1 has constraints, make sure
			// they are the same, otherwise error.
			pvConstraints := pathvars.Constraints(pv.Constraints).String()
			if pvConstraints != p.Constraints {
				// TODO Add regular error handling
				panic(fmt.Sprintf("Constraints mismatch: %s != %s", pvConstraints, p.Constraints))
			}
		}
	}
	// Now loop through all the path vars to see if we need to add any from the path
	// vars to the slice of APIParamV1.
	paramsMap = apiParams.Map()
	for _, pv := range pathVars {
		_, ok := paramsMap[string(pv.Name)]
		if ok {
			continue
		}
		apiParam := cfgldr.NewAPIParamV1(cfgldr.APIParamV1Args{
			NameSpec:    pv.String(),
			Type:        string(pv.Type.Slug()),
			Constraints: pathvars.Constraints(pv.Constraints).String(),
		})
		apiParams = append(apiParams, apiParam)
	}

	// Finally now loop through to collected and updated slice of APIParamV1 and
	// parse to convert to an Endpoint Param.
	for _, apiParam := range apiParams {
		var p EndpointParam
		p, err = ParseEndpointParam(apiParam.NameSpec, apiParam.Location, apiParam)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		epParams = append(epParams, p)
	}

	err = errors.Join(errs...)

end:
	return epParams, err
}

// ParseEndpointParam converts a configuration parameter into a validated
// EndpointParam. It handles both APIParamV1 and APIParamsMapValue formats,
// parsing the parameter specification and validating all constraints.
func ParseEndpointParam(nameSpec string, location string, cfg cfgldr.APIParam) (p EndpointParam, err error) {
	var props *pathvars.NameSpecProps
	var cc []pathvars.Constraint
	var dt pathvars.PVDataType
	param, ok := cfg.(cfgldr.APIParamV1)
	if !ok {
		paramSpec, ok := cfg.(cfgldr.APIParamsMapValue)
		if !ok {
			err = errors.Join(ErrInvalidAPIEndpointParameter, ErrCannotTypeAssert,
				fmt.Errorf("from_type=%T", cfg),
				fmt.Errorf("to_type=%T", (*cfgldr.APIParamV1)(nil)),
				fmt.Errorf("parameter_value=%v", cfg),
			)
			goto end
		}
		param, err = cfgldr.ParseAPIParamV1(nameSpec, string(paramSpec))
	}
	if err != nil {
		goto end
	}
	props, err = pathvars.ParseNameSpecProps(param.NameSpec)
	if err != nil {
		goto end
	}
	if props == nil {
		// Added this here because Goland flags props.DataType as possibly being null. I
		// don't see how it could be possible, but maybe Goland knows something I don't?
		panic(fmt.Sprintf("NameSpecProps are nil when err is also nil; spec=%s", param.NameSpec))
	}
	if props.DataType != nil {
		dt = *props.DataType
	}
	if param.Type != "" {
		dt, err = pathvars.ParsePVDataType(param.Type)
		if err != nil {
			goto end
		}
		cc, err = pathvars.ParseConstraints(param.Constraints, dt)
		if err != nil {
			goto end
		}
		p = NewEndpointParam(EndpointParamArgs{
			Props: *props,

			Type:        dt,
			Location:    pathvars.LocationType(location),
			Constraints: cc,
			RawValue:    param.String(),
		})
	}
end:
	return p, err
}

type Props = pathvars.NameSpecProps

type EndpointParams []EndpointParam

func (eps EndpointParams) FilterByNames(names []string) (out []EndpointParam) {
	var namesRegexp *regexp.Regexp
	if len(names) == 0 {
		goto end
	}
	out = make([]EndpointParam, len(eps))
	namesRegexp = regexp.MustCompile(fmt.Sprintf("^(%s)$", strings.Join(names, "|")))
	for i, ep := range eps {
		if !namesRegexp.MatchString(string(ep.Name)) {
			continue
		}
		out[i] = ep
	}
end:
	return out
}

// EndpointParam represents a parameter that can be extracted from HTTP requests
// and used in SQL query execution. Parameters can come from URL path segments,
// query strings, or JSON request bodies.
type EndpointParam struct {
	Props
	Type        pathvars.PVDataType   // Data type for validation and conversion
	Location    pathvars.LocationType // How the parameter is used (path, query, body, header)
	Constraints []pathvars.Constraint // Validation constraints (min/max, regex, etc.)
	nameSpec    pathvars.PVNameSpec
}

func (p EndpointParam) NameSpec() (ns pathvars.PVNameSpec) {
	if p.nameSpec == "" {
		p.nameSpec = pathvars.PVNameSpec(p.Props.String())
	}
	return p.nameSpec
}

func (p EndpointParam) RawValue() (s string) {
	return p.Props.RawValue
}
func (p EndpointParam) HasProps() bool {
	props := p.Props
	return props.Name != "" && props.RawValue != ""
}

func (p EndpointParam) String() (s string) {
	var sb strings.Builder
	var cs string
	if len(p.Constraints) != 0 {
		sb.WriteByte(':')
		for _, c := range p.Constraints {
			sb.WriteString(c.String())
			sb.WriteByte(',')
		}
		cs = sb.String()
		cs = cs[:len(cs)-1]
	}
	name := string(p.Props.Name)
	typ := string(p.Type.Slug())
	if cs == "" && name == typ {
		s = fmt.Sprintf("{%s}", name)
		goto end
	}
	s = fmt.Sprintf("{%s:%s%s}", name, typ, cs)
end:
	return s
}

// EndpointParamArgs contains the configuration for creating a new EndpointParam.
type EndpointParamArgs struct {
	Props
	Type        pathvars.PVDataType   // Parameter data type
	Location    pathvars.LocationType // Parameter usage type
	Constraints []pathvars.Constraint // Validation constraints
	RawValue    string
}

// NewEndpointParam creates a new EndpointParam with the specified name and configuration.
// If no constraints are provided, an empty slice is initialized.
func NewEndpointParam(args EndpointParamArgs) EndpointParam {
	if args.Constraints == nil {
		args.Constraints = make([]pathvars.Constraint, 0)
	}
	return EndpointParam{
		Props:       args.Props,
		Type:        args.Type,
		Location:    args.Location,
		Constraints: args.Constraints,
	}
}
