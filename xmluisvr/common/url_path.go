package common

import (
	"errors"
	"fmt"
	"regexp"
)

type URLPath string
type RelativeURL string

// urlPathRegexp allows for matching (maybe) method prefixed URL paths
//
//		Rationale for characters allowed:
//			a-z, A-Z: standard alphanumeric characters for resource names, actions, identifiers
//			0-9: numeric identifiers, version numbers, IDs
//			/ : path separators (required for all URL paths)
//			_ : underscore for multi-word resources like /user_profile, /order_history
//			. : file extensions like /config.json, /schema.xml, or version separators like /v1.2
//			{} for path parameters: /users/{id}
//			 : for namespaces like /api:v1
//	   ?: indicates start of query parameters
//	   &: separates query parameters
//			[] for array parameters like /users[0] or query-style paths
//			@ for special endpoints like /users/@me or /files/@latest
//			! for actions like /cache/!clear or /system/!restart
//			$ for special references like /users/$current or jQuery-style selectors
//			; for matrix parameters like /users;type=admin
//			= for inline parameters like /search=query
//			+ for some REST conventions: /users/search+filter
//			- : hyphen for kebab-case naming like /health-check, /api-docs, /user-settings
var relativeURLPathRegexpString = `[a-zA-Z0-9/_.{}:?&\[\]@!$;=+-]*`
var urlPathRegexp = regexp.MustCompile(`^(/?` + relativeURLPathRegexpString + `)$`)
var relativeURLPathRegexp = regexp.MustCompile(`^(` + relativeURLPathRegexpString + `)$`)

// ParseURLPath currently parses the FULL relative URL *TEMPLATE*, not just the path.
// TODO Rename this to ParseURLPathTemplate, or similar?
//
//	Move to pathvars package
func ParseURLPath(p string) (up URLPath, err error) {
	if p == "" {
		err = ErrURLPathMustNotBeEmpty
	}
	if !urlPathRegexp.MatchString(p) {
		err = errors.Join(ErrInvalidURLPath, fmt.Errorf("url_path=%s", p))
		goto end
	}
	up = URLPath(p)
end:
	return up, err
}

// ParseRelativeURL currently parses a (potentially partial) relative URL
// *TEMPLATE*, not just the path.
// TODO Rename this to ParseRelativeURLPathTemplate, or similar?
//
//	Move to pathvars package
func ParseRelativeURL(p string) (up RelativeURL, err error) {
	if p == "" {
		err = ErrURLPathMustNotBeEmpty
	}
	if p[0] == '/' {
		err = errors.Join(ErrURLPathMustNotBeginWithSlash,
			fmt.Errorf("url_path=%s", p),
		)
		goto end
	}
	if !relativeURLPathRegexp.MatchString(p) {
		err = errors.Join(ErrInvalidURLPath, fmt.Errorf("relative_url=%s", p))
		goto end
	}
	up = RelativeURL(p)
end:
	return up, err
}
