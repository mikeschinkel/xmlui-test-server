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

func (c *IntegerRangeConstraint) Rule() string {
	return fmt.Sprintf("%d..%d", c.min, c.max)
}

// ParseIntRangeConstraint parses min..max format for integers
func ParseIntRangeConstraint(rangeSpec string) (constraint *IntegerRangeConstraint, err error) {
	var parts []string
	var minimum, maximum int64
	var errs []error

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = errors.Join(
			ErrExpectedRangeFormat,
			fmt.Errorf("range=%s", rangeSpec),
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	minimum, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		err = errors.Join(
			ErrInvalidMinimumValue,
			fmt.Errorf("minimum=%s", parts[0]),
			err,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(parts) == 1 {
		err = errors.Join(errs...)
		goto end
	}

	maximum, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		err = errors.Join(
			ErrInvalidMaximumValue,
			fmt.Errorf("maximum=%s", parts[1]),
			err,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if minimum > maximum {
		err = errors.Join(
			ErrInvalidMinMaxValue,
			fmt.Errorf("minimum=%d", minimum),
			fmt.Errorf("maximum=%d", maximum),
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("range=%s", rangeSpec),
			errors.Join(errs...),
		)
		goto end
	}

	constraint = NewIntRangeConstraint(minimum, maximum)

end:
	return constraint, err
}
