package pathvars

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	RegisterConstraint(&IntegerRangeConstraint{})
}

var _ Constraint = (*IntegerRangeConstraint)(nil)

// IntegerRangeConstraint validates integer ranges
type IntegerRangeConstraint struct {
	baseConstraint
	min int64
	max int64
}

func NewIntRangeConstraint(min int64, max int64) *IntegerRangeConstraint {
	c := &IntegerRangeConstraint{min: min, max: max}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *IntegerRangeConstraint) ValidDateTypes() []PVDataType {
	return []PVDataType{IntegerType}
}

func (c *IntegerRangeConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseIntRangeConstraint(value)
}

func (c *IntegerRangeConstraint) Type() ConstraintType {
	return RangeConstraintType
}

func (c *IntegerRangeConstraint) Validate(value string) (err error) {
	var n int64

	n, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		goto end
	}

	if n < c.min || n > c.max {
		err = fmt.Errorf("value must be between %d and %d", c.min, c.max)
	}

end:
	return err
}

func (c *IntegerRangeConstraint) String() string {
	return fmt.Sprintf("%s[%d..%d]", c.Type(), c.min, c.max)
}

// ParseIntRangeConstraint parses min..max format for integers
func ParseIntRangeConstraint(rangeSpec string) (constraint *IntegerRangeConstraint, err error) {
	var parts []string
	var minimum, maximum int64

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("reason=%s", "expected format 'range[min..max]'"),
		)
		goto end
	}

	minimum, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%q", parts[0]),
			fmt.Errorf("reason=%s", "invalid minimum value"),
		)
		goto end
	}

	maximum, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("maximum=%q", parts[1]),
			fmt.Errorf("reason=%s", "invalid maximum value"),
		)
		goto end
	}

	if minimum > maximum {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%d", minimum),
			fmt.Errorf("maximum=%d", maximum),
			fmt.Errorf("reason=%s", "minimum value cannot be greater than maximum value"),
		)
		goto end
	}

	constraint = NewIntRangeConstraint(minimum, maximum)

end:
	return constraint, err
}
