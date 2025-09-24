// Package pathvars/route defines the Route type which represents a compiled HTTP route.
// Routes are created by the Router and contain the HTTP method, parsed template,
// and index information needed for request matching.
package pathvars

// Route represents a compiled HTTP endpoint with its method, template, and routing index.
// Routes are created during router compilation and used for efficient request matching.
type Route struct {
	// Method specifies the HTTP method for this route (GET, POST, etc.).
	// An empty string means the route matches any HTTP method.
	Method string

	// Template contains the parsed path template with parameters and regex for matching.
	Template *Template

	// Index indicates the position of this route in the router's route list.
	// This can be used to identify which specific route was matched.
	Index int
}
