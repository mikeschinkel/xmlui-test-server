package apipkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

// ParseEndpoints converts a slice of configuration endpoint definitions
// into parsed Endpoint structs. Each endpoint is validated during parsing.
func ParseEndpoints(cfgEPs []*cfgldr.APIEndpointV2) (eps []*Endpoint, err error) {
	var errs []error
	eps = make([]*Endpoint, len(cfgEPs))
	for i, cfgEP := range cfgEPs {
		eps[i], err = ParseEndpoint(cfgEP)
		errs = append(errs, err)
	}
	return eps, errors.Join(errs...)
}

// ErrOneOfQueryAndQueryFileMustNotBeEmpty is returned when an endpoint has neither
// an inline query nor a query file specified.
var ErrOneOfQueryAndQueryFileMustNotBeEmpty = errors.New("at least one of query for query file must not be empty")

// ParseEndpoint converts a configuration endpoint into a parsed Endpoint struct.
// It validates all fields including HTTP method, URL path, parameters, and SQL configuration.
func ParseEndpoint(cfg *cfgldr.APIEndpointV2) (ep *Endpoint, err error) {
	var errs []error

	ep = &Endpoint{
		Description: cfg.Description,
		Query:       common.QueryString(cfg.Query),
		QueryFile:   common.Filepath(cfg.QueryFile),
	}
	method, path := splitEndPoint(cfg.Endpoint)
	// TODO: Should we allow defining endpoints without explicitly specifying a method?
	ep.method, err = common.ParseHTTPMethod(method, common.EmptyOk)
	errs = append(errs, err)
	ep.path, err = common.ParseURLPath(path)
	errs = append(errs, err)
	ep.QueryFile, err = common.ParseFilepath(cfg.QueryFile)
	errs = append(errs, err)
	ep.Params, err = ParseParams(cfg.Params)
	errs = append(errs, err)
	ep.RowsExpected, err = common.ParseCardinality(cfg.Cardinality)
	errs = append(errs, err)
	ep.RowType, err = common.ParseDBRowType(cfg.RowType)
	errs = append(errs, err)
	ep.ColumnTypes, err = common.ParseColumnTypes(cfg.ColumnTypes)
	errs = append(errs, err)
	if ep.Query == "" && ep.QueryFile == "" {
		errs = append(errs, ErrOneOfQueryAndQueryFileMustNotBeEmpty)
	}
	err = errors.Join(errs...)
	if err != nil {
		*ep = Endpoint{}
		err = errors.Join(err, fmt.Errorf("endpoint=%s", cfg.Endpoint))
	}

	return ep, err
}

// EndPointString represents a string representation of an HTTP endpoint (e.g., "GET /users/:id").
type EndPointString string

// Endpoint represents a parsed API endpoint configuration with all validation complete.
// It contains the HTTP method, URL path, SQL query, parameters, and response formatting options.
type Endpoint struct {
	Params        []Param
	RowsExpected  common.Cardinality
	Description   string              // Human-readable description of the endpoint
	Query         common.QueryString  // Inline SQL query to execute
	QueryFile     common.Filepath     // Path to external SQL file (relative to config file)
	queryFilepath common.Filepath     // Resolved absolute path to SQL file
	RowType       common.DBRowType    // Format for returning results (json, columns, etc.)
	ColumnTypes   []common.DBDataType // Expected data types for result columns
	method        common.HTTPMethod   // HTTP method (GET, POST, etc.)
	path          common.URLPath      // URL path pattern with parameter placeholders
}

func (ep *Endpoint) PathVarsParameters() (params []pathvars.Parameter) {
// ParsePathVarsParameters converts endpoint parameters into pathvars.Parameter instances
// for use with the routing system. This enables path parameter extraction and validation.
	params = make([]pathvars.Parameter, 0, len(ep.Params))
	for _, p := range ep.Params {
		panic("FINISH THIS")
		param := pathvars.NewParameter(pathvars.ParameterArgs{
			Name:         string(p.Name),
			ParamType:    pathvars.UnspecifiedParameterType,
			DataType:     0,
			Constraints:  nil,
			Position:     0,
			Original:     "",
			MultiSegment: false,
			Optional:     false,
			DefaultValue: nil,
		})
		params = append(params, param)
	}
	return params
}

// GetQuery returns the SQL query for this endpoint, loading from a file if necessary.
// If QueryFile is specified, it loads the SQL from the file relative to the provided directory.
// Otherwise, it returns the inline Query string.
func (ep *Endpoint) GetQuery(dir common.DirPath) (q common.QueryString, queryFile common.Filepath, err error) {
	var queryBytes []byte

	q = ep.Query

	// Check if SQL should be loaded from a file
	if ep.QueryFile == "" {
		// Use the inline SQL from the APIConfig definition
		goto end
	}
	if ep.queryFilepath != "" {
		goto end
	}
	// Determine the APIConfig description file's directory to make relative paths work

	// Build the SQL file path relative to the APIConfig description file
	ep.queryFilepath = common.Filepath(filepath.Join(string(dir), string(ep.QueryFile)))

	// Read the SQL file
	queryBytes, err = os.ReadFile(string(queryFile))
	if err != nil {
		err = errors.Join(ErrFailedToReadQueryFile, err)
		goto end
	}

	q = common.QueryString(queryBytes)
end:
	return q, ep.queryFilepath, err
}

// Endpoint returns a string representation of the endpoint in "METHOD /path" format.
func (ep *Endpoint) Endpoint() EndPointString {
	return EndPointString(fmt.Sprintf("%s %s",
		strings.ToUpper(string(ep.method)),
		ep.path),
	)
}

// Path returns the URL path pattern for this endpoint.
func (ep *Endpoint) Path() common.URLPath {
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
		path = string(ep)
		goto end
	}
end:
	return method, path
}
