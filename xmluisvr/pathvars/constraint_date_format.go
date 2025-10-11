package pathvars

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func init() {
	RegisterConstraint(&DateFormatConstraint{})
}

var _ Constraint = (*DateFormatConstraint)(nil)

// DateFormatConstraint validates date formats
type DateFormatConstraint struct {
	baseConstraint
	format string
	parser func(string) (time.Time, error)
}

func NewDateFormatConstraint(format string, parser func(string) (time.Time, error)) *DateFormatConstraint {
	c := &DateFormatConstraint{format: format, parser: parser}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *DateFormatConstraint) ValidDateTypes() []PVDataType {
	return []PVDataType{DateType, UUIDType, StringType}
}

func (c *DateFormatConstraint) Type() ConstraintType {
	return FormatConstraintType
}

func (c *DateFormatConstraint) Parse(value string, dataType PVDataType) (ct Constraint, err error) {
	// Handle different data types for format constraints
	switch dataType {
	case DateType:
		ct, err = ParseDateFormatConstraint(value)
	case UUIDType:
		ct, err = ParseUUIDFormatConstraint(value)
	case StringType:
		// Check if this is a UUID-like format for strings
		switch strings.ToLower(value) {
		case "ulid", "ksuid", "nanoid":
			ct, err = ParseUUIDFormatConstraint(value)
		default:
			err = errors.Join(
				ErrInvalidConstraint,
				ErrStringFormatOnlySupportsIDFormats,
				fmt.Errorf("value=%s", value),
				fmt.Errorf("data_type=%v", dataType),
			)
		}
	default:
		err = errors.Join(
			ErrInvalidConstraint,
			ErrFormatConstraintUnsupportedDataType,
			fmt.Errorf("data_type=%v", dataType),
		)
	}
	return ct, err
}

func (c *DateFormatConstraint) Validate(value string) (err error) {
	_, err = c.parser(value)
	if err != nil {
		// For partial dates, try to validate against partial formats
		err = c.validatePartialDate(value)
		if err != nil {
			err = fmt.Errorf("invalid date format, expected %s or partial match", c.format)
		}
	}
	return err
}

// validatePartialDate validates partial date formats for multi-segment parameters
func (c *DateFormatConstraint) validatePartialDate(value string) (err error) {
	var partialLayout string

	// Extract tokens from the original format to see what partial formats are valid
	partialLayout, err = c.buildPartialLayout(value)
	if err != nil {
		goto end
	}

	_, err = time.Parse(partialLayout, value)
	if err != nil {
		err = fmt.Errorf("partial date validation failed")
	}

end:
	return err
}

// buildPartialLayout creates a Go time layout for partial date validation
func (c *DateFormatConstraint) buildPartialLayout(value string) (layout string, err error) {
	var segments []string
	var formatSegments []string
	var partialFormat string

	// Split both the value and format by slashes to see how many segments we have
	segments = strings.Split(value, "/")
	formatSegments = strings.Split(c.format, "/")

	// Build a partial format using only the segments we have
	if len(segments) > len(formatSegments) {
		err = errors.Join(
			ErrInvalidConstraint,
			ErrMoreSegmentsThanFormat,
			fmt.Errorf("value=%s", value),
			fmt.Errorf("format=%s", c.format),
			fmt.Errorf("value_segments=%d", len(segments)),
			fmt.Errorf("format_segments=%d", len(formatSegments)),
		)
		goto end
	}

	// Take only the format segments we need for the actual value segments
	partialFormat = strings.Join(formatSegments[:len(segments)], "/")

	// Use the existing token-based parser to build the Go time layout
	layout, err = buildGoTimeLayout(partialFormat)
	if err != nil {
		err = errors.Join(
			ErrFailedToBuildPartialLayout,
			fmt.Errorf("value=%s", value),
			fmt.Errorf("format=%s", c.format),
			fmt.Errorf("partial_format=%s", partialFormat),
			err,
		)
	}

end:
	return layout, err
}

func (c *DateFormatConstraint) Rule() string {
	return c.format
}

// ParseDateFormatConstraint parses date format specifications
func ParseDateFormatConstraint(spec string) (constraint *DateFormatConstraint, err error) {
	var goLayout string
	var parser func(string) (time.Time, error)

	// Handle special built-in formats first
	if spec == "iso8601" {
		parser = func(s string) (time.Time, error) {
			return time.Parse(time.RFC3339, s)
		}
		constraint = NewDateFormatConstraint(spec, parser)
		goto end
	}

	// ParseBytes the format specification to build Go time layout
	goLayout, err = buildGoTimeLayout(spec)
	if err != nil {
		err = errors.Join(
			ErrInvalidDateFormatSpec,
			fmt.Errorf("spec=%s", spec),
			err,
		)
		goto end
	}

	parser = func(s string) (time.Time, error) {
		return time.Parse(goLayout, s)
	}

	constraint = NewDateFormatConstraint(spec, parser)

end:
	return constraint, err
}

// buildGoTimeLayout converts a date format specification to Go time layout
func buildGoTimeLayout(spec string) (layout string, err error) {
	var result []rune
	var i int
	var hasHour bool
	var token string
	var hasAnyToken bool

	// ParseBytes character by character, looking for date/time tokens
	for i < len(spec) {
		// Try to match each possible token at current position
		token, hasHour, err = matchToken(spec, i, hasHour)
		if err != nil {
			goto end
		}

		if token != "" {
			// Found a token, append its Go layout equivalent
			result = append(result, []rune(token)...)
			i += tokenLength(spec, i)
			hasAnyToken = true
		} else {
			// Not a token, append the literal character
			result = append(result, rune(spec[i]))
			i++
		}
	}

	// Validate that we found at least one date/time token
	if !hasAnyToken {
		err = errors.Join(
			ErrInvalidConstraint,
			ErrNoValidDateTimeTokens,
			fmt.Errorf("spec=%s", spec),
		)
		goto end
	}

	layout = string(result)

end:
	return layout, err
}

// matchToken attempts to match a date/time token at the given position
func matchToken(spec string, pos int, hasHour bool) (goToken string, newHasHour bool, err error) {
	newHasHour = hasHour

	// Check for each possible token
	if matchesAt(spec, pos, "yyyy") {
		goToken = "2006"
		goto end
	}
	if matchesAt(spec, pos, "yy") {
		goToken = "06"
		goto end
	}
	if matchesAt(spec, pos, "dd") {
		goToken = "02"
		goto end
	}
	if matchesAt(spec, pos, "hh") {
		goToken = "15"
		newHasHour = true
		goto end
	}
	if matchesAt(spec, pos, "ss") {
		goToken = "05"
		goto end
	}
	if matchesAt(spec, pos, "ii") {
		// ii always means minutes
		goToken = "04"
		goto end
	}
	if matchesAt(spec, pos, "mm") {
		// mm disambiguation: month if no hour seen yet, minutes if hour seen
		switch {
		case hasHour:
			goToken = "04" // minutes
		case isStandaloneMM(spec):
			// Check if this is a standalone mm (ambiguous)
			err = errors.Join(
				ErrInvalidConstraint,
				ErrAmbiguousMMToken,
				fmt.Errorf("spec=%s", spec),
				fmt.Errorf("position=%d", pos),
			)
			goto end
		default:
			goToken = "01" // month
		}
		goto end
	}

end:
	return goToken, newHasHour, err
}

// matchesAt checks if the given token matches at the specified position
func matchesAt(spec string, pos int, token string) bool {
	var i int

	if pos+len(token) > len(spec) {
		return false
	}

	for i = 0; i < len(token); i++ {
		if spec[pos+i] != token[i] {
			return false
		}
	}

	return true
}

// tokenLength returns the length of the token at the given position
func tokenLength(spec string, pos int) int {
	if matchesAt(spec, pos, "yyyy") {
		return 4
	}
	if matchesAt(spec, pos, "yy") || matchesAt(spec, pos, "mm") ||
		matchesAt(spec, pos, "dd") || matchesAt(spec, pos, "hh") ||
		matchesAt(spec, pos, "ii") || matchesAt(spec, pos, "ss") {
		return 2
	}
	return 1
}

// isStandaloneMM checks if the spec contains only "mm" as a token (ambiguous case)
func isStandaloneMM(spec string) bool {
	var tokenCount int
	var i int

	// Count tokens in the spec
	for i < len(spec) {
		if matchesAt(spec, i, "yyyy") {
			tokenCount++
			i += 4
		} else if matchesAt(spec, i, "yy") || matchesAt(spec, i, "mm") ||
			matchesAt(spec, i, "dd") || matchesAt(spec, i, "hh") ||
			matchesAt(spec, i, "ii") || matchesAt(spec, i, "ss") {
			tokenCount++
			i += 2
		} else {
			i++
		}
	}

	// If only one token and it's mm, it's standalone and ambiguous
	return tokenCount == 1 && hasOnlyMMToken(spec)
}

// hasOnlyMMToken checks if the spec contains only the mm token
func hasOnlyMMToken(spec string) bool {
	var i int

	for i < len(spec) {
		if matchesAt(spec, i, "mm") {
			// Found mm, continue to check if there are other tokens
			i += 2
			continue
		}
		if matchesAt(spec, i, "yyyy") || matchesAt(spec, i, "yy") ||
			matchesAt(spec, i, "dd") || matchesAt(spec, i, "hh") ||
			matchesAt(spec, i, "ii") || matchesAt(spec, i, "ss") {
			// Found another token
			return false
		}
		i++
	}

	// Check if we actually found mm in the spec
	return containsToken(spec, "mm")
}

// containsToken checks if the spec contains the given token
func containsToken(spec string, token string) bool {
	var i int

	for i < len(spec) {
		if matchesAt(spec, i, token) {
			return true
		}
		i++
	}
	return false
}

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

func (c *DateRangeConstraint) ValidDateTypes() []PVDataType {
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
		err = errors.Join(ErrDateLessThanMinimum,
			fmt.Errorf("minimum_date=%s", c.min.Format(time.DateOnly)),
		)
		goto end
	}

	if d.After(c.max) {
		err = errors.Join(ErrDateGreaterThanMaximum,
			fmt.Errorf("maximum_date=%s", c.max.Format(time.DateOnly)),
		)
		goto end
	}

end:
	if err != nil {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("date_value=%s", value),
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
		err = errors.Join(
			ErrInvalidConstraint, ErrExpectedRangeFormat,
		)
		goto end
	}

	// ParseBytes minimum date (try ISO format first)
	minimum, err = time.Parse(time.DateOnly, parts[0])
	if err != nil {
		err = errors.Join(ErrInvalidMinimumValue, ErrExpectedISO8601DateFormat,
			fmt.Errorf("minimum=%s", parts[0]),
			err,
		)
		goto end
	}

	// ParseBytes maximum date (try ISO format first)
	maximum, err = time.Parse(time.DateOnly, parts[1])
	if err != nil {
		err = errors.Join(ErrInvalidMaximumValue, ErrExpectedISO8601DateFormat,
			fmt.Errorf("maximum=%s", parts[1]),
			err,
		)
		goto end
	}

	if minimum.After(maximum) {
		err = errors.Join(
			ErrInvalidConstraint, ErrInvalidMinMaxDate,
			fmt.Errorf("minimum=%s", minimum.Format(time.DateOnly)),
			fmt.Errorf("maximum=%s", maximum.Format(time.DateOnly)),
		)
		goto end
	}

	constraint = NewDateRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = errors.Join(
			fmt.Errorf("range=%s", rangeSpec),
			err,
		)
	}
	return constraint, err
}
