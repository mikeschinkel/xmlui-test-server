package common

import (
	"errors"
	"fmt"
	"regexp"
)

type URLPath string

// urlPathRegexp allows for matching (maybe) method prefixed URL paths
//
//	Rationale for characters allowed:
//		a-z, A-Z: standard alphanumeric characters for resource names, actions, identifiers
//		0-9: numeric identifiers, version numbers, IDs
//		/ : path separators (required for all URL paths)
//		_ : underscore for multi-word resources like /user_profile, /order_history
//		. : file extensions like /config.json, /schema.xml, or version separators like /v1.2
//		{} for path parameters: /users/{id}
//		 : for namespaces like /api:v1
//		[] for array parameters like /users[0] or query-style paths
//		@ for special endpoints like /users/@me or /files/@latest
//		! for actions like /cache/!clear or /system/!restart
//		$ for special references like /users/$current or jQuery-style selectors
//		; for matrix parameters like /users;type=admin
//		= for inline parameters like /search=query
//		+ for some REST conventions: /users/search+filter
//		- : hyphen for kebab-case naming like /health-check, /api-docs, /user-settings
var urlPathRegexp = regexp.MustCompile(`^(/[a-zA-Z0-9/_.{}:[\]@!$;=+-]*)$`)

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
