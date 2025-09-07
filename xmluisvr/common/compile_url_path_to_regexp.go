package common

import (
	"fmt"
	"regexp"
)

// ParseURLPathToRegexp sonverts a path template to a regexp
// Example: "/clients/:id" -> "^/clients/([^/]+)$"
func CompileURLPathToRegexp(path URLPath) (*regexp.Regexp, error) {
	// Escape any special regexp characters in the path
	escaped := regexp.QuoteMeta(string(path))

	// Replace :paramName with a capturing group
	re := regexp.MustCompile(`:([^/]+)`)
	regexpPath := re.ReplaceAllString(escaped, "([^/]+)")

	// Add start and end anchors
	return regexp.Compile(fmt.Sprintf("^%s$", regexpPath))
}
