package pathvars

import (
	"fmt"
	"strings"
	"time"
)

func init() {
	RegisterConstraint(&DateRangeConstraint{})
}

var _ Constraint = (*DateRangeConstraint)(nil)

// DateRangeConstraint validates date ranges
type DateRangeConstraint struct {
	baseConstraint
	min time.Time
	max time.Time
}

func NewDateRangeConstraint(min time.Time, max time.Time) *DateRangeConstraint {
	c := &DateRangeConstraint{
		min: min,
		max: max,
	}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *DateRangeConstraint) ValidDataTypes() []PVDataType {
	return []PVDataType{DateType}
}

func (c *DateRangeConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseDateRangeConstraint(value)
}

func (c *DateRangeConstraint) Type() ConstraintType {
	return RangeConstraintType
}

func (c *DateRangeConstraint) Validate(value string) (err error) {
	var d time.Time

	// Try common date formats
	formats := []string{
		time.DateOnly, //  "2006-01-02"
		"01/02/2006",
		"02/01/2006",
		time.DateTime,     //  "2006-01-02 15:04:05"
		time.RFC3339[:20], //  "2006-01-02T15:04:05Z"
		time.RFC3339[:19], //  "2006-01-02T15:04:05"
		time.RFC3339,      //  "2006-01-02T15:04:05Z07:00"
		time.ANSIC,        //  "Mon Jan _2 15:04:05 2006"
		time.UnixDate,     //  "Mon Jan _2 15:04:05 MST 2006"
		time.RubyDate,     //  "Mon Jan 02 15:04:05 -0700 2006"
		time.RFC822,       //  "02 Jan 06 15:04 MST"
		time.RFC822Z,      //  "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
		time.RFC850,       //  "Monday, 02-Jan-06 15:04:05 MST"
		time.RFC1123,      //  "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z,     //  "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
		time.RFC3339Nano,  //  "2006-01-02T15:04:05.999999999Z07:00"
	}

	for _, format := range formats {
		d, err = time.Parse(format, value)
		if err == nil {
			break
		}
	}

	if err != nil {
		err = ErrInvalidDateFormat
		goto end
	}
	err = nil

	if d.Before(c.min) {
		err = NewErr(ErrDateLessThanMinimum,
			"minimum_date", c.min.Format(time.DateOnly),
		)
		goto end
	}

	if d.After(c.max) {
		err = NewErr(ErrDateGreaterThanMaximum,
			"maximum_date", c.max.Format(time.DateOnly),
		)
		goto end
	}

end:
	if err != nil {
		err = NewErr(
			"date_value", value,
			err,
		)
	}
	return err
}

func (c *DateRangeConstraint) Rule() string {
	return fmt.Sprintf("%s..%s", c.min.Format(time.DateOnly), c.max.Format(time.DateOnly))
}

// ParseDateRangeConstraint parses min..max format for dates
func ParseDateRangeConstraint(rangeSpec string) (constraint *DateRangeConstraint, err error) {
	var parts []string
	var minimum, maximum time.Time

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = NewErr(ErrExpectedRangeFormat)
		goto end
	}

	// ParseBytes minimum date (try dateonly format first)
	minimum, err = time.Parse(time.DateOnly, parts[0])
	if err != nil {
		err = NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	// ParseBytes maximum date (try dateonly format first)
	maximum, err = time.Parse(time.DateOnly, parts[1])
	if err != nil {
		err = NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum.After(maximum) {
		err = NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMinMaxDate,
			"minimum", minimum.Format(time.DateOnly),
			"maximum", maximum.Format(time.DateOnly),
		)
		goto end
	}

	constraint = NewDateRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = WithErr(err,
			ErrInvalidRangeConstraint,
			"range_spec", rangeSpec,
		)
	}
	return constraint, err
}
