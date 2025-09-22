package pathvars

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type PathSpec string
type Method string
type Path string

// Router holds compiled routes
type Router struct {
	routes    []*Route
	compiled  bool
	maxParams int
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// AddRoute adds a route to the router
func (r *Router) AddRoute(pathSpec PathSpec, params []Parameter) (err error) {
	return r.AddRouteWithIndex(pathSpec, params, len(r.routes))
}

// AddRouteWithIndex adds a route to the router with the specified index
func (r *Router) AddRouteWithIndex(pathSpec PathSpec, params []Parameter, index int) (err error) {
	var method string
	var path string
	var template *Template
	var route *Route
	var paramCount int

	// Parse "GET /users/{id}" format
	method, path, err = ParsePathSpec(pathSpec)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("path_spec=%q", pathSpec),
		)
		goto end
	}

	template, err = ParseTemplate(path)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("path_spec=%q", pathSpec),
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

	for _, param := range params {
		template.params[param.name] = &param
	}

	// Track max params for optimization
	paramCount = len(template.params)
	if paramCount > r.maxParams {
		r.maxParams = paramCount
	}

	route = &Route{
		Method:   method,
		Template: template,
		Index:    index,
	}

	r.routes = append(r.routes, route)

end:
	return err
}

// Compile pre-compiles all routes
func (r *Router) Compile() (err error) {
	// Validate all routes are properly configured
	// Set compiled flag
	r.compiled = true
	return err
}

// Match matches a request against routes
func (r *Router) Match(req *http.Request) (result MatchResult, err error) {
	var varsMap VarsMap
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
		if route.Method != "" && route.Method != req.Method {
			continue
		}

		varsMap, matched = route.Template.Match(u.Path, u.RawQuery)
		if matched {
			result = MatchResult{
				Index:   route.Index,
				varsMap: varsMap,
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

// ParsePathSpec splits "GET /path" into method and path
func ParsePathSpec(spec PathSpec) (method string, path string, err error) {
	var found bool

	// Parse the method and path from spec
	// Handle cases: "GET /path", "/path" (any method)
	if spec == "" {
		err = errors.Join(
			ErrInvalidTemplate,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("reason=%s", "empty path specification"),
		)
		goto end
	}

	// If starts with '/', it's just a path (any method)
	if spec[0] == '/' {
		path = string(spec)
		goto end
	}

	// Otherwise, split on first space
	method, path, found = strings.Cut(string(spec), " ")
	if !found {
		err = errors.Join(
			ErrInvalidTemplate,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("method=%q", method),
			fmt.Errorf("reason=%s", "missing space between method and path"),
		)
		goto end
	}

	// Validate path starts with '/'
	if path == "" || path[0] != '/' {
		err = errors.Join(
			ErrInvalidTemplate,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("method=%q", method),
			fmt.Errorf("path=%q", path),
			fmt.Errorf("reason=%s", "path must start with '/' or be non-empty"),
		)
		goto end
	}

end:
	return method, path, err
}
