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

func (c *LengthConstraint) String() string {
	return fmt.Sprintf("%s[%d..%d]", c.Type(), c.min, c.max)
}

// ParseLengthConstraint parses min..max format
func ParseLengthConstraint(rangeSpec string) (constraint *LengthConstraint, err error) {
	var parts []string
	var minimum, maximum int

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("reason=%s", "expected format 'min..max'"),
		)
		goto end
	}

	minimum, err = strconv.Atoi(parts[0])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%q", parts[0]),
			fmt.Errorf("reason=%s", "invalid minimum length"),
		)
		goto end
	}

	maximum, err = strconv.Atoi(parts[1])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("maximum=%q", parts[1]),
			fmt.Errorf("reason=%s", "invalid maximum length"),
		)
		goto end
	}

	if minimum > maximum {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%d", minimum),
			fmt.Errorf("maximum=%d", maximum),
			fmt.Errorf("reason=%s", "invalid length range (min > max)"),
		)
		goto end
	}

	if minimum < 0 {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%d", minimum),
			fmt.Errorf("reason=%s", "invalid length range (min < 0)"),
		)
		goto end
	}

	constraint = NewLengthConstraint(minimum, maximum)

end:
	return constraint, err
}
