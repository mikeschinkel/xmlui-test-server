package pathvars

import (
	"fmt"
	"regexp"
	"strings"
)

func init() {
	RegisterConstraint(&RegexConstraint{})
}

var _ Constraint = (*RegexConstraint)(nil)

// RegexConstraint validates against regex regex
type RegexConstraint struct {
	baseConstraint
	regex *regexp.Regexp
	raw   string
}

func NewRegexConstraint(regex *regexp.Regexp, raw string) *RegexConstraint {
	c := &RegexConstraint{
		regex: regex,
		raw:   raw,
	}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *RegexConstraint) ValidDataTypes() []PVDataType {
	return []PVDataType{
		StringType,
		SlugType,
		AlphanumericType,
		EmailType,
		IdentifierType,
	}
}

func (c *RegexConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseRegexConstraint(value)
}

func (c *RegexConstraint) Type() ConstraintType {
	return RegexConstraintType
}

func (c *RegexConstraint) Validate(value string) (err error) {
	if !c.regex.MatchString(value) {
		err = fmt.Errorf("value does not match regex %s", c.raw)
	}
	return err
}

func (c *RegexConstraint) Rule() string {
	return c.raw
}

func (c *RegexConstraint) ErrorSuggestion(param *Parameter, value, example string) string {
	// TODO Add more specific advice, and don't use this advice when not applicable
	return "Do not include ^ or $ anchors in your regex pattern; regex patterns automatically match the full parameter value."
}

// ParseRegexConstraint parses a regex pattern and automatically anchors it for full string matching.
// Patterns must not include ^ or $ anchors - they are added automatically to ensure the pattern
// matches the complete parameter value, not just a substring.
func ParseRegexConstraint(pattern string) (constraint *RegexConstraint, err error) {
	var regex *regexp.Regexp
	var anchoredPattern string
	var errs []error
	var hasStart, hasEnd bool

	if pattern == "" {
		err = NewErr(ErrEmptyRegexPattern)
		goto end
	}

	// Check for anchors and collect all errors before returning
	hasStart = strings.HasPrefix(pattern, "^")
	hasEnd = strings.HasSuffix(pattern, "$")

	if hasStart && hasEnd {
		errs = append(errs, ErrRegexPatternContainsBothAnchors)
	}
	if hasStart {
		errs = append(errs, ErrRegexPatternContainsStartAnchor)
	}
	if hasEnd {
		errs = append(errs, ErrRegexPatternContainsEndAnchor)
	}

	if len(errs) > 0 {
		err = NewErr(
			ErrInvalidRegexPattern,
			CombineErrs(errs),
		)
		goto end
	}

	// Auto-anchor the pattern for full string matching
	anchoredPattern = "^" + pattern + "$"

	// Compile the anchored pattern
	regex, err = regexp.Compile(anchoredPattern)
	if err != nil {
		err = NewErr(
			ErrInvalidRegexPattern,
			err,
		)
		goto end
	}

	// Store original pattern (without anchors) for display
	constraint = &RegexConstraint{regex: regex, raw: pattern}

end:
	if err != nil {
		err = WithErr(err,
			ErrInvalidRegexConstraint,
			"regex", pattern,
		)
	}
	return constraint, err
}
