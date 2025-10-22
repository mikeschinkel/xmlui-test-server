package pathvars

import (
	"fmt"
	"strconv"
	"strings"
)

func init() {
	RegisterConstraint(&LengthConstraint{})
}

var _ Constraint = (*LengthConstraint)(nil)

// LengthConstraint validates string length
type LengthConstraint struct {
	baseConstraint
	min int
	max int
}

func NewLengthConstraint(min int, max int) *LengthConstraint {
	c := &LengthConstraint{min: min, max: max}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *LengthConstraint) ValidDataTypes() []PVDataType {
	return []PVDataType{
		StringType,
		IdentifierType,
		AlphanumericType,
		SlugType,
		EmailType,
	}
}

func (c *LengthConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseLengthConstraint(value)
}

func (c *LengthConstraint) Type() ConstraintType {
	return LengthConstraintType
}

func (c *LengthConstraint) Validate(value string) (err error) {
	length := len(value)
	if length < c.min || length > c.max {
		err = fmt.Errorf("length must be between %d and %d", c.min, c.max)
	}
	return err
}

func (c *LengthConstraint) Rule() string {
	return fmt.Sprintf("%d..%d", c.min, c.max)
}

// ParseLengthConstraint parses min..max format
func ParseLengthConstraint(lengthSpec string) (constraint *LengthConstraint, err error) {
	var parts []string
	var minimum, maximum int

	// Split by ".."
	parts = strings.Split(lengthSpec, "..")
	if len(parts) != 2 {
		err = NewErr(ErrExpectedLengthFormat)
		goto end
	}

	minimum, err = strconv.Atoi(parts[0])
	if err != nil {
		err = NewErr(ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	maximum, err = strconv.Atoi(parts[1])
	if err != nil {
		err = NewErr(ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum > maximum {
		err = NewErr(
			ErrInvalidLengthRangeMinGreaterThanMax,
			"minimum", minimum,
			"maximum", maximum,
		)
		goto end
	}

	if minimum < 0 {
		err = NewErr(
			ErrInvalidLengthRangeNegativeMin,
			"minimum", minimum,
		)
		goto end
	}

	constraint = NewLengthConstraint(minimum, maximum)

end:
	if err != nil {
		err = WithErr(err,
			"length_spec", lengthSpec,
		)
	}
	return constraint, err
}
