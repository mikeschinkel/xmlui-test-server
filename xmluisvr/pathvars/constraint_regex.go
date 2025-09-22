package pathvars

import (
	"errors"
	"fmt"
	"regexp"
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

func (c *RegexConstraint) ValidDateTypes() []PVDataType {
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

func (c *RegexConstraint) String() string {
	return fmt.Sprintf("%s[%s]", c.Type(), c.raw)
}

// ParseRegexConstraint parses a regex pattern
func ParseRegexConstraint(pattern string) (constraint *RegexConstraint, err error) {
	var regex *regexp.Regexp

	if pattern == "" {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("pattern=%q", pattern),
			fmt.Errorf("reason=%s", "empty regex pattern"),
		)
		goto end
	}

	regex, err = regexp.Compile(pattern)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("pattern=%q", pattern),
			fmt.Errorf("reason=%s", "invalid regular expression"),
		)
		goto end
	}

	constraint = &RegexConstraint{regex: regex, raw: pattern}

end:
	return constraint, err
}
