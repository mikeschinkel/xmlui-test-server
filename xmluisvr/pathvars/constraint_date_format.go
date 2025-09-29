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

func (c *DateFormatConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	// Handle different data types for format constraints
	switch dataType {
	case DateType:
		return ParseDateFormatConstraint(value)
	case UUIDType:
		return ParseUUIDFormatConstraint(value)
	case StringType:
		// Check if this is a UUID-like format for strings
		switch strings.ToLower(value) {
		case "ulid", "ksuid", "nanoid":
			return ParseUUIDFormatConstraint(value)
		default:
			return nil, errors.Join(
				ErrInvalidConstraint,
				fmt.Errorf("value=%q", value),
				fmt.Errorf("data_type=%v", dataType),
				fmt.Errorf("reason=%s", "string format constraint only supports ulid, ksuid, nanoid"),
			)
		}
	default:
		return nil, errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("data_type=%v", dataType),
			fmt.Errorf("reason=%s", "format constraint only supports date, uuid, and string data types"),
		)
	}
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
			fmt.Errorf("value=%q", value),
			fmt.Errorf("format=%q", c.format),
			fmt.Errorf("valueSegments=%d", len(segments)),
			fmt.Errorf("formatSegments=%d", len(formatSegments)),
			fmt.Errorf("reason=%s", "more segments in value than in format"),
		)
		goto end
	}

	// Take only the format segments we need for the actual value segments
	partialFormat = strings.Join(formatSegments[:len(segments)], "/")

	// Use the existing token-based parser to build the Go time layout
	layout, err = buildGoTimeLayout(partialFormat)
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("value=%q", value),
			fmt.Errorf("format=%q", c.format),
			fmt.Errorf("partialFormat=%q", partialFormat),
			fmt.Errorf("reason=%s", "failed to build partial layout"),
		)
	}

end:
	return layout, err
}

func (c *DateFormatConstraint) String() string {
	return fmt.Sprintf("%s[%s]", c.Type(), c.format)
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
			err,
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("reason=%s", "invalid date format specification"),
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
			fmt.Errorf("spec=%q", spec),
			fmt.Errorf("reason=%s", "no valid date/time tokens found in format"),
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
				fmt.Errorf("spec=%q", spec),
				fmt.Errorf("position=%d", pos),
				fmt.Errorf("reason=%s", "ambiguous 'mm' token - use 'ii' for minutes or add other tokens for context"),
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
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"01/02/2006",
		"02/01/2006",
	}

	for _, format := range formats {
		d, err = time.Parse(format, value)
		if err == nil {
			break
		}
	}

	if err != nil {
		err = fmt.Errorf("invalid date format: %s", value)
		goto end
	}

	if d.Before(c.min) || d.After(c.max) {
		err = fmt.Errorf("date must be between %s and %s", c.min.Format("2006-01-02"), c.max.Format("2006-01-02"))
	}

end:
	return err
}

func (c *DateRangeConstraint) String() string {
	return fmt.Sprintf("%s[%s..%s]", RangeConstraintType, c.min.Format("2006-01-02"), c.max.Format("2006-01-02"))
}

// ParseDateRangeConstraint parses min..max format for dates
func ParseDateRangeConstraint(rangeSpec string) (constraint *DateRangeConstraint, err error) {
	var parts []string
	var minimum, maximum time.Time

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

	// ParseBytes minimum date (try ISO format first)
	minimum, err = time.Parse("2006-01-02", parts[0])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%q", parts[0]),
			fmt.Errorf("reason=%s", "invalid minimum date (expected YYYY-MM-DD format)"),
		)
		goto end
	}

	// ParseBytes maximum date (try ISO format first)
	maximum, err = time.Parse("2006-01-02", parts[1])
	if err != nil {
		err = errors.Join(
			err,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("maximum=%q", parts[1]),
			fmt.Errorf("reason=%s", "invalid maximum date (expected YYYY-MM-DD format)"),
		)
		goto end
	}

	if minimum.After(maximum) {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("rangeSpec=%q", rangeSpec),
			fmt.Errorf("minimum=%s", minimum.Format("2006-01-02")),
			fmt.Errorf("maximum=%s", maximum.Format("2006-01-02")),
			fmt.Errorf("reason=%s", "minimum date cannot be after maximum date"),
		)
		goto end
	}

	constraint = NewDateRangeConstraint(minimum, maximum)

end:
	return constraint, err
}
