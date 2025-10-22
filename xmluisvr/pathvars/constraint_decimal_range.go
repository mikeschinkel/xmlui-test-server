package pathvars

import (
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

func (c *DecimalRangeConstraint) ValidDataTypes() []PVDataType {
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

func (c *DecimalRangeConstraint) Rule() string {
	return fmt.Sprintf("%g..%g", c.min, c.max)
}

// ParseDecimalRangeConstraint parses min..max format for decimals
func ParseDecimalRangeConstraint(rangeSpec string) (constraint *DecimalRangeConstraint, err error) {
	var parts []string
	var minimum, maximum float64

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = NewErr(ErrExpectedRangeFormat)
		goto end
	}

	minimum, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		err = NewErr(
			ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	maximum, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		err = NewErr(
			ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum > maximum {
		err = NewErr(
			ErrInvalidMinMaxValue,
			fmt.Errorf("minimum=%g", minimum),
			fmt.Errorf("maximum=%g", maximum),
		)
		goto end
	}

	constraint = NewDecimalRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = WithErr(err,
			ErrInvalidRangeValue,
			"range_spec", rangeSpec,
		)
	}
	return constraint, err
}
