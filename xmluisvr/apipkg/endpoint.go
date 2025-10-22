package apipkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/jsonxtractr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"

	. "github.com/xmlui-org/xmlui-test-server/xmluisvr/doterr"
)

// ParseEndpoints converts a slice of configuration endpoint definitions
// into parsed Endpoint structs. Each endpoint is validated during parsing.
func ParseEndpoints(cfgEPs []*cfgldr.APIEndpointV2, basePath common.URLPath, db dbpkg.Database) (eps []*Endpoint, err error) {
	var errs []error
	eps = make([]*Endpoint, len(cfgEPs))
	for i, cfgEP := range cfgEPs {
		eps[i], err = ParseEndpoint(cfgEP, basePath, db)
		if err != nil {
			errs = append(errs, NewErr(
				ErrInvalidAPIEndpointParameter,
				err,
			))
		}
	}
	err = CombineErrs(errs)
	if err != nil {
		err = NewErr(
			ErrParsingFailed,
			ErrParsingOfMultipleEndpointsFailed,
			CombineErrs(errs),
		)
	}
	return eps, err
}

func ParseQuery(query common.QueryString, db dbpkg.Database) (pq dbqvars.ParsedQuery, err error) {

	// QueryFileExt is a proxy for Query Type.
	// TODO: Maybe add a first-class QueryType later
	qt := strings.ToLower(db.QueryFileExt())
	switch qt {
	case ".sql":
		pq, err = dbqvars.ParseSQL(dbqvars.SQLQuery(query), db.GetFormatParamFunc())
	default:
		err = NewErr(ErrQueryTypeParsingNotYetSupported, "query_type", qt)
	}
	return pq, err
}

// ParseEndpoint converts a configuration endpoint into a parsed Endpoint struct.
// It validates all fields including HTTP method, URL path, parameters, and SQL configuration.
func ParseEndpoint(cfg *cfgldr.APIEndpointV2, basePath common.URLPath, db dbpkg.Database) (ep *Endpoint, err error) {
	var errs []error
	var q string
	ep = &Endpoint{
		Description: cfg.Description,
	}
	q, err = cfg.GetQuery()
	if err == nil {
		ep.ParsedQuery, err = ParseQuery(common.QueryString(q), db)
	}
	errs = AppendErr(errs, err)

	// TODO: Allow or disallow defining endpoints without explicitly specifying a method? Maybe we should require "ANY"?
	ep.method, err = common.ParseHTTPMethod(cfg.Method, common.EmptyOk)
	errs = AppendErr(errs, err)
	var relPath *pathvars.ParsedTemplate
	relPath, err = pathvars.ParseTemplate(cfg.Path)
	errs = AppendErr(errs, err)

	// TODO Allow or disallow root-based URLs that ignore basepath; which to choose?
	//      Need override setting to explicitly allow
	// 			UNTIL THEN, we strip any leading slash (`/`)
	relPath.Normalize()

	ep.path = pathvars.Template(fmt.Sprintf("%s/%s", basePath, relPath))
	ep.Params, err = ParseEndpointParams(cfg.Params, ep.path)
	errs = AppendErr(errs, err)
	ep.pathParsed = true
	ep.Cardinality, err = dbqvars.ParseCardinality(cfg.Cardinality)
	errs = AppendErr(errs, err)
	ep.RowType, err = dbqvars.ParseDBRowType(cfg.RowType)
	errs = AppendErr(errs, err)
	ep.ColumnTypes, err = dbqvars.ParseColumnTypes(cfg.ColumnTypes)
	errs = AppendErr(errs, err)
	err = CombineErrs(errs)
	if err != nil {
		ep = nil
		err = NewErr(
			ErrParsingFailed,
			ErrEndpointParsingFailed,
			err,
			"endpoint", cfg.Endpoint(),
		)
	}
	return ep, err
}

var (
	ErrParsingFailed                    = errors.New("parsing failed")
	ErrEndpointParsingFailed            = errors.New("endpoint parsing failed")
	ErrParsingOfMultipleEndpointsFailed = errors.New("parsing of multiple endpoints failed")
)

// EndPointString represents a string representation of an HTTP endpoint (e.g., "GET /users/:id").
type EndPointString string

// Endpoint represents a parsed API endpoint configuration with all validation complete.
// It contains the HTTP method, URL path, SQL query, parameters, and response formatting options.
type Endpoint struct {
	Description   string               // Human-readable description of the endpoint
	ParsedQuery   dbqvars.ParsedQuery  // Query to execute parsed by dbqvars.ParseBytes()
	queryFilepath common.Filepath      // Resolved absolute path to SQL file
	Params        []EndpointParam      // Parameters that can be extracted from requests
	Cardinality   dbqvars.Cardinality  // Expected number of result rows (one, many, etc.)
	RowType       dbqvars.DBRowType    // Format for returning results (json, columns, etc.)
	ColumnTypes   []dbqvars.DBDataType // Expected data types for result columns
	method        common.HTTPMethod    // HTTP method (GET, POST, etc.)
	path          pathvars.Template    // URL path pattern with parameter placeholders
	pathParsed    bool
}

func (ep *Endpoint) GetBodyValuesMap(r io.Reader, selectors []jsonxtractr.Selector) (valuesMap jsonxtractr.ValuesMap, notFound []jsonxtractr.Selector, err error) {
	dbq := ep.ParsedQuery
	// Get the pathValuesMap needed for the SQL query from the URL path and query variables
	valuesMap, notFound, err = jsonxtractr.ExtractValuesFromReader(r, selectors)
	if errors.Is(err, jsonxtractr.ErrJSONValueSelectorCannotBeEmpty) {
		err = nil
		goto end
	}
	if err != nil {
		// TODO: Replace with an Response
		err = NewErr(ErrExtractingFromReader,
			err,
			"endpoint", ep.Endpoint(),
			"sql_query", dbq.QueryString(),
			"sql_params", dbq.Parameters(),
			"body_matched", valuesMap,
			"not_matched", notFound,
		)
		goto end
	}
end:
	return valuesMap, notFound, err
}

type ParameterValuesArgs struct {
	ValuesMap  pathvars.ValuesMap
	BodyReader io.Reader
	Headers    http.Header // TODO: Not yet supported
	Database   dbpkg.Database
}

func (ep *Endpoint) parameterTypeMap() (tm map[string]pathvars.PVDataType) {
	// Create a map of parameter names to their types for quick lookup
	tm = make(map[string]pathvars.PVDataType)
	for _, param := range ep.Params {
		tm[string(param.Name)] = param.Type
	}
	return tm
}

func (ep *Endpoint) GetParameterValues(args ParameterValuesArgs) (queryValues []any, missing []apiresp.MissingParameter, err error) {
	var pathValuesMap pathvars.ValuesMap
	var namesNotFound []pathvars.Identifier
	var jsonValuesMap jsonxtractr.ValuesMap
	var notFound []jsonxtractr.Selector
	var selectors []dbqvars.Selector
	var epParams []EndpointParam

	dbq := ep.ParsedQuery
	parameters := dbq.Parameters()
	occurrences := dbq.Occurrences()

	// Get the pathValuesMap needed for the SQL query from the URL path and query variables
	ids := pathvars.Identifiers(parameters.Identifiers())
	pathValuesMap, namesNotFound = args.ValuesMap.GetValues(ids)

	// Get the selectors to search JSON - only dotted selectors (e.g., task.title) should
	// be extracted from the body. Query parameters and path parameters are already in pathValuesMap.
	selectors = parameters.DottedSelectors()
	// NOTE: We do NOT add namesNotFound to selectors here because those are simple parameter names
	// (not dotted) that should come from path/query, not the JSON body.

	switch {
	case args.BodyReader != nil:
		// We got a reader for the JSON body - extract dotted body parameters
		jsonValuesMap, notFound, err = ep.GetBodyValuesMap(args.BodyReader, jsonxtractr.ToSelectors(selectors))
		if err != nil {
			err = NewErr(ErrExtractingJSONBodyValues, err)
			goto end
		}
		// Combine path/query params not found with body params not found
		notFound = combineStringsAsY(namesNotFound, notFound)
	default:
		// We did NOT get a reader for the JSON body
		// Any parameters not found in path/query are missing
		notFound = combineStringsAsY(namesNotFound, []jsonxtractr.Selector{})
	}

	// Build queryValues array based on ALL occurrences (including duplicates)
	// For SQL binding, we need one value per placeholder, even if the same parameter appears multiple times
	queryValues = make([]any, len(occurrences))
	for i, token := range occurrences {
		qv, ok := pathValuesMap.Get(pathvars.Identifier(token.Name))
		if ok {
			queryValues[i] = qv
			continue
		}
		if args.BodyReader == nil {
			// We did not get a body ready so no jsonValuesMap to look at
			continue
		}
		qv, ok = jsonValuesMap[jsonxtractr.Selector(token.Name)]
		if ok {
			queryValues[i] = qv
			continue
		}
	}

	// Convert values based on parameter types for SQL compatibility
	queryValues = ep.convertValuesForSQL(occurrences.Parameters(), queryValues, args)

	// Build list of missing parameters for error reporting
	epParams = EndpointParams(ep.Params).FilterByNames(jsonxtractr.Selectors(notFound).Strings())
	missing = make([]apiresp.MissingParameter, len(epParams))
	for i, p := range epParams {
		missing[i] = apiresp.MissingParameter{
			// TODO: Converting an Identifier to a Selector. p.Name should probably be a Selector
			Selector: rfc9457.Selector(p.Name),
			Location: apiresp.LocationType(p.Location),
			Expected: "", // TODO Can we populate this?
			Received: "", // TODO Can we populate this?
			Message:  "", // TODO Can we populate this?
		}
	}
end:
	return queryValues, missing, err
}

// ParsePathVarParameters converts endpoint parameters into pathvars.Parameter instances
// for use with the routing system. This enables path parameter extraction and validation.
func (ep *Endpoint) ParsePathVarParameters() (params []pathvars.Parameter, err error) {
	var errs []error
	params = make([]pathvars.Parameter, 0, len(ep.Params))
	for i, p := range ep.Params {
		var dt pathvars.PVDataType
		props := p.Props
		if props.DataType != nil {
			dt = *props.DataType
		}
		if dt == pathvars.UnspecifiedDataType {
			dt, err = pathvars.ParseParameterDataType(string(props.Name), string(p.Type.Slug()))
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if p.RawValue() == "" {
			// TODO Remove this after we ensure RawValue is set
			panic("PARAMETER RAW VALUE NOT SET")
		}
		params = append(params, pathvars.NewParameter(pathvars.ParameterArgs{
			Position:    i,
			NameProps:   props,
			Location:    p.Location,
			DataType:    dt,
			Constraints: p.Constraints,
			Original:    p.RawValue(),
		}))
	}
	return params, err
}

//// GetQuery returns the SQL query for this endpoint, loading from a file if necessary.
//// If QueryFile is specified, it loads the SQL from the file relative to the provided directory.
//// Otherwise, it returns the inline Query string.
//func (ep *Endpoint) GetQuery(dir common.DirPath) (q common.QueryString, queryFile common.Filepath, err error) {
//	panic("FIX THIS")
//	return q, "", err
//}

// Endpoint returns a string representation of the endpoint in "METHOD /path" format.
func (ep *Endpoint) Endpoint() EndPointString {
	return EndPointString(fmt.Sprintf("%s %s",
		strings.ToUpper(string(ep.method)),
		ep.path),
	)
}

// Path returns the URL path pattern for this endpoint.
func (ep *Endpoint) Path() pathvars.Template {
	return ep.path
}

// Method returns the HTTP method for this endpoint, with ANY method converted to empty string.
func (ep *Endpoint) Method() (m common.HTTPMethod) {
	m = ep.RawMethod()
	if m == common.ANYMethod {
		m = ""
		goto end
	}
end:
	return m
}

// RawMethod returns the raw HTTP method without any processing.
func (ep *Endpoint) RawMethod() common.HTTPMethod {
	return ep.method
}

// splitEndPoint separates an endpoint string like "GET /path" into method and path components.
// If no method is specified, the entire string is treated as the path.
func splitEndPoint(ep string) (method, path string) {
	method, path, found := strings.Cut(ep, " ")
	if !found {
		path = ep
		goto end
	}
end:
	return method, path
}

var (
	ErrConvertingValuesForSQL = errors.New("converting values for SQL")
)

// convertValuesForSQL converts parameter values to SQL-compatible types.
// Currently handles boolean to integer conversion for SQLite compatibility.
func (ep *Endpoint) convertValuesForSQL(parameters dbqvars.Parameters, values []any, args ParameterValuesArgs) (_ []any) {
	var converted []any

	// Create a map of parameter names to their types for quick lookup
	tm := ep.parameterTypeMap()

	// Convert values based on their parameter types
	converted = make([]any, len(values))
	for i, param := range parameters {
		value := values[i]
		if value == nil {
			converted[i] = value
			continue
		}

		// Look up the parameter type
		dt, exists := tm[string(param.Name)]
		if !exists {
			// If type not found, keep original value
			converted[i] = value
			continue
		}
		// Convert if needed
		converted[i] = convertValueForSQL(value, dt, args)
	}
	return converted
}

func convertValueForSQL(value any, dt pathvars.PVDataType, args ParameterValuesArgs) any {
	switch dt {
	case pathvars.BooleanType:
		value = args.Database.ConvertValue(value, dbqvars.IntegerDBDataType)
	}
	return value
}
