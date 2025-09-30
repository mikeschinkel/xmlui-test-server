// Package pathvars provides path variable routing and parameter extraction functionality
// for HTTP requests. It supports typed path parameters with validation constraints,
// query parameter handling, and regex-based route matching.
//
// The package enables parsing URL templates like "/users/{id:int:range[1..1000]}"
// and matching them against incoming HTTP requests, extracting and validating
// parameter values according to specified data types and constraints.
//
// Key features:
//   - Typed path parameters (string, int, uuid, date, etc.)
//   - Parameter validation constraints (range, length, regex, enum, etc.)
//   - Multi-segment parameters for capturing multiple path segments
//   - Optional parameters with default values
//   - Query parameter extraction and validation
//   - Efficient regex-based route matching
//
// Example usage:
//
//	router := pathvars.NewRouter()
//	params := []pathvars.Parameter{
//		// Parameter definitions go here
//	}
//	err := router.AddRoute("GET" "/users/{id:int}", params)
//	if err != nil {
//		// handle error
//	}
//	err = router.Compile()
//	if err != nil {
//		// handle error
//	}
//
//	// Later, during request handling:
//	result, err := router.Match(request)
//	if err == nil {
//		userId, found := result.GetValue("id")
//		// use the extracted parameter
//	}
package pathvars

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// PathSpec represents a path specification string like "GET /users/{id}" or "/users/{id}".
type PathSpec string

// Method represents an HTTP method string like "GET", "POST", etc.
type Method string

// Path represents a URL path string like "/users/{id}".
type Path string

// Router holds compiled routes and provides request matching functionality.
// A Router must be compiled before it can match requests.
type Router struct {
	routes    []*Route
	compiled  bool
	maxParams int
}

// NewRouter creates a new router instance.
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

type RouteArgs struct {
	Parameters  []Parameter
	Index       int
	Description string              // Human-readable description of the endpoint
	Cardinality common.Cardinality  // Expected number of result rows (one, many, etc.)
	RowType     common.DBRowType    // Format for returning results (json, columns, etc.)
	ColumnTypes []common.DBDataType // Expected data types for result columns
}

// AddRoute adds a route to the router with the specified path specification and parameters.
// The pathSpec can be in format "METHOD /path" (e.g., "GET /users/{id}") or just "/path"
// for any method. Parameters define the expected path and query parameters for this route.
func (r *Router) AddRoute(method common.HTTPMethod, path common.URLPath, args *RouteArgs) (err error) {
	var template *Template
	var route *Route
	var paramCount int

	if args == nil {
		args = &RouteArgs{}
	}

	if path == "" {
		// Trim leading slash ('/') on sub path
		path = "/"
	}
	if path[0] != '/' {
		// Trim leading slash ('/') on sub path
		path = "/" + path
	}

	template, err = ParseTemplate(path)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("path_spec=%q", path),
			fmt.Errorf("method=%q", method),
			fmt.Errorf("path=%q", path),
		)
		goto end
	}

	if template == nil {
		// This if statement if only here because without it Goland is reporting that
		// `template` might be nil in the expressions below even though I traced through
		// the logic and it cannot be nil if err==nil.
		goto end
	}

	if len(args.Parameters) != 0 {
		for _, param := range args.Parameters {
			template.params[param.Name] = param
		}

		// Track max params for optimization
		paramCount = len(template.params)
		if paramCount > r.maxParams {
			r.maxParams = paramCount
		}
	}

	if args.Index == 0 {
		args.Index = len(r.routes) + 1
	}
	route = &Route{
		Method:      method,
		Template:    template,
		Index:       args.Index,
		Description: args.Description,
		Cardinality: args.Cardinality,
		RowType:     args.RowType,
		ColumnTypes: args.ColumnTypes,
	}

	r.routes = append(r.routes, route)

end:
	return err
}

// Compile pre-compiles all routes for efficient matching.
// This must be called before using Match() method.
func (r *Router) Compile() (err error) {
	// Validate all routes are properly configured
	// Set compiled flag
	//panic("IMPLEMENT ME!")
	r.compiled = true
	return err
}

// Match matches an HTTP request against the compiled routes and returns
// the first matching route along with extracted parameter values.
// Returns ErrAPIRouterNotCompiled if the router hasn't been compiled,
// or ErrNoMatch if no route matches the request.
func (r *Router) Match(req *http.Request) (result MatchResult, err error) {
	var valuesMap ValuesMap
	var matched bool

	u := req.URL

	if !r.compiled {
		err = errors.Join(
			ErrAPIRouterNotCompiled,
			fmt.Errorf("method=%q", req.Method),
			fmt.Errorf("path=%q", u.Path),
			fmt.Errorf("query_string=%q", u.RawQuery),
			fmt.Errorf("route_count=%d", len(r.routes)),
			fmt.Errorf("reason=%s", "router must be compiled before matching"),
		)
		goto end
	}

	for _, route := range r.routes {
		// Check method match (empty method means any)
		if route.Method != "" && route.Method != common.HTTPMethod(req.Method) {
			continue
		}

		valuesMap, matched = route.Template.Match(u.Path, u.RawQuery)
		if matched {
			result = MatchResult{
				Index:     route.Index,
				Route:     route,
				valuesMap: valuesMap,
			}
			goto end
		}
	}

	err = errors.Join(
		ErrNoMatch,
		fmt.Errorf("method=%q", req.Method),
		fmt.Errorf("path=%q", u.Path),
		fmt.Errorf("query_string=%q", u.RawQuery),
		fmt.Errorf("route_count=%d", len(r.routes)),
		fmt.Errorf("reason=%s", "no route matched the request"),
	)

end:
	return result, err
}
