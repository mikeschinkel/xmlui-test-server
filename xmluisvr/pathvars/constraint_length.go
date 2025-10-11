package pathvars

import (
	"errors"
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

func (c *LengthConstraint) ValidDateTypes() []PVDataType {
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
func ParseLengthConstraint(rangeSpec string) (constraint *LengthConstraint, err error) {
	var parts []string
	var minimum, maximum int

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = errors.Join(
			ErrInvalidConstraint, ErrExpectedRangeFormat,
		)
		goto end
	}

	minimum, err = strconv.Atoi(parts[0])
	if err != nil {
		err = errors.Join(ErrInvalidMinimumValue,
			fmt.Errorf("minimum=%s", parts[0]),
			err,
		)
		goto end
	}

	maximum, err = strconv.Atoi(parts[1])
	if err != nil {
		err = errors.Join(ErrInvalidMaximumValue,
			fmt.Errorf("maximum=%s", parts[1]),
			err,
		)
		goto end
	}

	if minimum > maximum {
		err = errors.Join(
			ErrInvalidConstraint,
			ErrInvalidLengthRangeMinGreaterThanMax,
			fmt.Errorf("minimum=%d", minimum),
			fmt.Errorf("maximum=%d", maximum),
		)
		goto end
	}

	if minimum < 0 {
		err = errors.Join(
			ErrInvalidConstraint,
			ErrInvalidLengthRangeNegativeMin,
			fmt.Errorf("minimum=%d", minimum),
		)
		goto end
	}

	constraint = NewLengthConstraint(minimum, maximum)

end:
	if err != nil {
		err = errors.Join(
			fmt.Errorf("range=%s", rangeSpec),
			err,
		)
	}
	return constraint, err
}
