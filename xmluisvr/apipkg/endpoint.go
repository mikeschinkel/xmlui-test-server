package apipkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/jsonutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// ParseEndpoints converts a slice of configuration endpoint definitions
// into parsed Endpoint structs. Each endpoint is validated during parsing.
func ParseEndpoints(cfgEPs []*cfgldr.APIEndpointV2, basePath common.URLPath, db dbpkg.Database) (eps []*Endpoint, err error) {
	var errs []error
	eps = make([]*Endpoint, len(cfgEPs))
	for i, cfgEP := range cfgEPs {
		eps[i], err = ParseEndpoint(cfgEP, basePath, db)
		if err != nil {
			errs = append(errs, errors.Join(
				ErrInvalidAPIEndpointParameter,
				err,
			))
		}
	}
	return eps, errors.Join(errs...)
}

func ParseQuery(query common.QueryString, db dbpkg.Database) (pq dbqvars.ParsedQuery, err error) {

	// QueryFileExt is a proxy for Query Type.
	// TODO: Maybe add a first-class QueryType later
	qt := strings.ToLower(db.QueryFileExt())
	switch qt {
	case ".sql":
		pq, err = dbqvars.ParseSQL(dbqvars.SQLQuery(query), db.GetFormatParamFunc())
	default:
		err = errors.Join(ErrQueryTypeParsingNotYetSupported, fmt.Errorf("query_type=%s", qt))
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
	errs = append(errs, err)

	// TODO: Allow or disallow defining endpoints without explicitly specifying a method? Maybe we should require "ANY"?
	ep.method, err = common.ParseHTTPMethod(cfg.Method, common.EmptyOk)
	errs = append(errs, err)
	var relPath *pathvars.ParsedTemplate
	relPath, err = pathvars.ParseTemplate(cfg.Path)
	errs = append(errs, err)

	// TODO Allow or disallow root-based URLs that ignore basepath; which to choose?
	//      Need override setting to explicitly allow
	ep.path = pathvars.Template(fmt.Sprintf("%s/%s", basePath, relPath))
	ep.Params, err = ParseEndpointParams(cfg.Params, ep.path)
	errs = append(errs, err)
	ep.pathParsed = true
	ep.Cardinality, err = dbqvars.ParseCardinality(cfg.Cardinality)
	errs = append(errs, err)
	ep.RowType, err = dbqvars.ParseDBRowType(cfg.RowType)
	errs = append(errs, err)
	ep.ColumnTypes, err = dbqvars.ParseColumnTypes(cfg.ColumnTypes)
	errs = append(errs, err)
	err = errors.Join(errs...)
	if err != nil {
		ep = nil
		err = errors.Join(err,
			fmt.Errorf("endpoint=%s", cfg.Endpoint()),
		)
	}

	return ep, err
}

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

func (ep *Endpoint) GetBodyValuesMap(r io.Reader, selectors []jsonutil.Selector) (valuesMap jsonutil.ValuesMap, notFound []jsonutil.Selector, err error) {
	dbq := ep.ParsedQuery
	// Get the pathValuesMap needed for the SQL query from the URL path and query variables
	valuesMap, notFound, err = jsonutil.ExtractValuesFromReader(r, selectors)
	if errors.Is(err, jsonutil.ErrJSONValueSelectorCannotBeEmpty) {
		err = nil
		goto end
	}
	if err != nil {
		// TODO: Replace with an Response
		err = errors.Join(ErrExtractingFromReader,
			err,
			fmt.Errorf("endpoint=%v", ep.Endpoint()),
			fmt.Errorf("sql_query=%v", dbq.QueryString()),
			fmt.Errorf("sql_params=%v", dbq.Parameters()),
			fmt.Errorf("body_matched=%v", valuesMap),
			fmt.Errorf("not_matched=%v", notFound),
		)
		goto end
	}
end:
	return valuesMap, notFound, err
}

type ParameterValuesSource struct {
	ValuesMap  pathvars.ValuesMap
	BodyReader io.Reader
	Headers    http.Header // TODO: Not yet supported
}

func (ep *Endpoint) GetParameterValues(pvs ParameterValuesSource) (queryValues []any, missing []apiutil.MissingParameter, err error) {
	var pathValuesMap pathvars.ValuesMap
	var namesNotFound []pathvars.Identifier
	var jsonValuesMap jsonutil.ValuesMap
	var notFound []jsonutil.Selector
	var selectors []dbqvars.Selector
	var epParams []EndpointParam

	dbq := ep.ParsedQuery
	parameters := dbq.Parameters()

	// Get the pathValuesMap needed for the SQL query from the URL path and query variables
	ids := pathvars.Identifiers(parameters.Identifiers())
	pathValuesMap, namesNotFound = pvs.ValuesMap.GetValues(ids)

	// Note get the selectors to search JSON
	selectors = parameters.DottedSelectors()
	if len(namesNotFound) > 0 {
		// If some non-dotted selectors were not found in path or query, add to selectors
		// to potentially extract values for from the JSON body.
		selectors = combineStringsAsY(namesNotFound, selectors)
	}

	switch {
	case pvs.BodyReader != nil:
		// We got a reader for the JSON body
		jsonValuesMap, notFound, err = ep.GetBodyValuesMap(pvs.BodyReader, jsonutil.ToSelectors(selectors))
		if err != nil {
			err = errors.Join(ErrExtractingJSONBodyValues, err)
			goto end
		}
	default:
		// We did NOT get a reader for the JSON body
		// Convert slice of []common.Identifier to slice of []common.Selector{}.
		notFound = combineStringsAsY(namesNotFound, []jsonutil.Selector{})
	}

	queryValues = make([]any, len(parameters))
	for i, p := range parameters {
		qv, ok := pathValuesMap[pathvars.Identifier(p.Name)]
		if ok {
			queryValues[i] = qv
			continue
		}
		if pvs.BodyReader == nil {
			// We did not get a body ready so no jsonValuesMap to look at
			continue
		}
		qv, ok = jsonValuesMap[jsonutil.Selector(p.Name)]
		if ok {
			queryValues[i] = qv
			continue
		}
	}
	missing = make([]apiutil.MissingParameter, len(notFound))
	epParams = EndpointParams(ep.Params).FilterByNames(jsonutil.Selectors(notFound).Strings())
	for i, p := range epParams {
		missing[i] = apiutil.MissingParameter{
			// TODO: Converting an Identifier to a Selector. p.Name should probably be a Selector
			Selector: rfc9457.Selector(p.Name),
			Location: apiutil.LocationType(p.Location),
			Expected: "", // TODO Can we populate this?
			Received: "", // TODO Can we populate this?
			Message:  "", // TODO Can we populate this?
		}
	}
end:
	return queryValues, missing, err
}

// ParsePathVarsParameters converts endpoint parameters into pathvars.Parameter instances
// for use with the routing system. This enables path parameter extraction and validation.
func (ep *Endpoint) ParsePathVarsParameters() (params []pathvars.Parameter, err error) {
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
			Location:    pathvars.LocationType(p.Location),
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
