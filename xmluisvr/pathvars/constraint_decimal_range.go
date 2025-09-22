package pathvars

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	RegisterConstraint(&DecimalRangeConstraint{})
}

var _ Constraint = (*DecimalRangeConstraint)(nil)

// DecimalRangeConstraint validates decimal ranges
type DecimalRangeConstraint struct {
	baseConstraint
	min float64
	max float64
}

func NewDecimalRangeConstraint(min float64, max float64) *DecimalRangeConstraint {
	c := &DecimalRangeConstraint{
		min: min,
		max: max,
	}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *DecimalRangeConstraint) ValidDateTypes() []PVDataType {
	return []PVDataType{DecimalType, RealType}
}

func (c *DecimalRangeConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseDecimalRangeConstraint(value)
}

func (c *DecimalRangeConstraint) Type() ConstraintType {
	return RangeConstraintType
}

func (c *DecimalRangeConstraint) Validate(value string) (err error) {
	var n float64

	n, err = strconv.ParseFloat(value, 64)
	if err != nil {
		goto end
	}

	if n < c.min || n > c.max {
		err = fmt.Errorf("value must be between %g and %g", c.min, c.max)
	}

end:
	return err
}

func (c *DecimalRangeConstraint) String() string {
	return fmt.Sprintf("%s[%g..%g]", c.Type(), c.min, c.max)
}

// ParseDecimalRangeConstraint parses min..max format for decimals
func ParseDecimalRangeConstraint(rangeSpec string) (constraint *DecimalRangeConstraint, err error) {
	var parts []string
	var minimum, maximum float64

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

	minimum, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%q", parts[0]),
			fmt.Errorf("reason=%s", "invalid minimum value"),
		)
		goto end
	}

	maximum, err = strconv.ParseFloat(parts[1], 64)
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
			fmt.Errorf("minimum=%g", minimum),
			fmt.Errorf("maximum=%g", maximum),
			fmt.Errorf("reason=%s", "minimum value cannot be greater than maximum value"),
		)
		goto end
	}

	constraint = NewDecimalRangeConstraint(minimum, maximum)

end:
	return constraint, err
}
