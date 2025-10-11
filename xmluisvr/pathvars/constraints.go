// Package pathvars/constraints defines the constraint system for parameter validation.
// Constraints provide additional validation rules beyond basic data type checking,
// such as ranges, formats, enums, and regular expressions.
package pathvars

import (
	"errors"
	"fmt"
	"strings"
)

// ConstraintType represents the type of constraint applied to a parameter.
type ConstraintType string

// Supported constraint types for parameter validation.
const (
	// FormatConstraintType validates parameter values against specific formats (e.g., date formats, UUID versions).
	FormatConstraintType ConstraintType = "format"

	// EnumConstraintType validates that parameter values match one of a predefined set of allowed values.
	EnumConstraintType ConstraintType = "enum"

	// LengthConstraintType validates that string parameter values fall within specified length ranges.
	LengthConstraintType ConstraintType = "length"

	// NotEmptyConstraintType validates that parameter values are not empty strings.
	NotEmptyConstraintType ConstraintType = "notempty"

	// RangeConstraintType validates that numeric parameter values fall within specified numeric ranges.
	RangeConstraintType ConstraintType = "range"

	// RegexConstraintType validates parameter values against regular expression patterns.
	RegexConstraintType ConstraintType = "regex"
)

type Constraints []Constraint

func (c Constraints) String() (s string) {
	sb := strings.Builder{}
	for _, constraint := range c {
		sb.WriteString(constraint.String())
		sb.WriteByte(',')
	}
	s = sb.String()
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

// Constraint interface defines the contract for parameter validation constraints.
// All constraint implementations must provide validation, parsing, and metadata methods.
type Constraint interface {
	// Validate checks if the given value satisfies this constraint.
	Validate(value string) error

	// String returns a human-readable representation of this constraint.
	String() string

	// Rule returns the constraint's rule from within square brackets
	Rule() string

	// Type returns the type of this constraint.
	Type() ConstraintType

	// Parse creates a new instance of this constraint from a string specification.
	Parse(value string, dataType PVDataType) (Constraint, error)

	// ValidDateTypes returns the data types that this constraint can be applied to.
	ValidDateTypes() []PVDataType

	// MapKey generates a unique key for constraint registry lookup.
	MapKey(dt PVDataTypeSlug) ConstraintMapKey

	// EnsureBaseConstraint sets up the base constraint relationship for proper functioning.
	EnsureBaseConstraint(Constraint)
}

// baseConstraint provides common functionality for all constraint implementations.
// It maintains a reference to the owning constraint for proper method delegation.
type baseConstraint struct {
	// owner holds a reference to the constraint that embeds this base.
	owner Constraint
}

// newBaseConstraint creates a new base constraint with the specified owner.
func newBaseConstraint(owner Constraint) baseConstraint {
	return baseConstraint{
		owner: owner,
	}
}

func (c *baseConstraint) String() string {
	return fmt.Sprintf("%s[%s]", c.owner.Type(), c.owner.Rule())
}

// EnsureBaseConstraint sets the owner reference for proper constraint operation.
func (c *baseConstraint) EnsureBaseConstraint(owner Constraint) {
	c.owner = owner
}

// MapKey generates a constraint registry key using the owner's type and data type.
func (c *baseConstraint) MapKey(dt PVDataTypeSlug) ConstraintMapKey {
	return GetConstraintMapKey(c.owner.Type(), dt)
}

// ParseConstraints parses constraint specifications from a string.
//
// ParseBytes constraint specs like:
//   - notempty
//   - range[0..100]
//   - length[5..50]
//   - regex[^[0-9]+$]
//   - enum[val1,val2,val3]
//   - For dates: format[iso8601], format[yyyy-mm-dd], etc.
//   - Multiple constraints: regex[^[0-9]+$],length[3..10]
func ParseConstraints(spec string, dataType PVDataType) (constraints []Constraint, err error) {
	var ctm ConstraintsMap
	var ct ConstraintType
	var constraint Constraint
	var pos, last, valueStart, constraintStart int
	var value string
	var errs []error
	var mode byte
	var ok bool
	var regexStart, regexEnd int

	typeName := dataType.Slug()
	const (
		typeMode  = 't'
		valueMode = 'v'
	)

	if spec == "" {
		goto end
	}

	ctm = GetConstraintsMap()

	// Pre-scan for regex constraint to find its true boundaries
	regexStart, regexEnd = findRegexBoundaries(spec)

	mode = typeMode
	last = len(spec)
	constraintStart = 0
	for pos < last {
		ch := spec[pos]
		pos++
		if isWhitespace(ch) {
			continue
		}

		//goland:noinspection GoDfaConstantCondition
		switch mode {
		case typeMode:
			if ch == '[' {
				// Found start of constraint value
				valueStart = pos
				ct = ConstraintType(spec[constraintStart : pos-1])
				key := GetConstraintMapKey(ct, typeName)
				constraint, ok = ctm[key]
				if !ok {
					errs = append(errs,
						errors.Join(ErrUnknownConstraintType,
							fmt.Errorf("constraint_spec=%s", spec),
							fmt.Errorf("constraint_type=%s", ct),
							fmt.Errorf("data_type=%s", typeName),
						),
					)
					continue
				}
				mode = valueMode
				continue
			}
			if ch == ',' || pos == last {
				// Found end of constraint type without brackets (like "notempty")
				var constraintEnd int
				constraintEnd = pos
				if ch == ',' {
					constraintEnd--
				}
				ct = ConstraintType(spec[constraintStart:constraintEnd])
				key := GetConstraintMapKey(ct, typeName)
				constraint, ok = ctm[key]
				if !ok {
					errs = append(errs,
						errors.Join(ErrUnknownConstraintType,
							fmt.Errorf("constraint_spec=%s", spec),
							fmt.Errorf("constraint_type=%s", ct),
							fmt.Errorf("data_type=%s", typeName),
						),
					)
					continue
				}
				// ParseBytes constraint with empty value (no arguments)
				constraint, err = constraint.Parse("", dataType)
				if err != nil {
					errs = append(errs,
						errors.Join(ErrParseFailed,
							fmt.Errorf("constraint_value=%s", ""),
							fmt.Errorf("constraint_type=%s", ct),
							fmt.Errorf("constraint_spec=%s", spec),
							fmt.Errorf("data_type=%s", typeName),
							err,
						),
					)
				} else {
					constraints = append(constraints, constraint)
				}

				if ch == ',' {
					// Skip past the comma and any whitespace to continue parsing next constraint
					for pos < last && (spec[pos] == ',' || isWhitespace(spec[pos])) {
						pos++
					}
					constraintStart = pos
					continue
				}
				break
			}
			if !isConstraintTypeChar(ch) {
				errs = append(errs,
					errors.Join(ErrInvalidSyntax,
						ErrInvalidConstraintTypeCharacter,
						fmt.Errorf("position=%d", pos),
						fmt.Errorf("character=%s", string(ch)),
						fmt.Errorf("constraint_type=%s", ct),
						fmt.Errorf("constraint_spec=%s", spec),
						fmt.Errorf("data_type=%s", typeName),
					),
				)
				continue
			}

		case valueMode:
			// Special handling for regex constraint - use pre-scanned boundaries
			if ct == RegexConstraintType && regexStart != -1 && constraintStart == regexStart {
				// Jump to the pre-scanned end position
				pos = regexEnd + 1
				value = spec[valueStart:regexEnd]
				constraint, err = constraint.Parse(value, dataType)
				if err != nil {
					errs = append(errs,
						errors.Join(ErrParseFailed,
							fmt.Errorf("constraint_value=%s", value),
							fmt.Errorf("constraint_type=%s", ct),
							fmt.Errorf("constraint_spec=%s", spec),
							fmt.Errorf("data_type=%s", dataType.Slug()),
							fmt.Errorf("start_pos=%d", valueStart),
							fmt.Errorf("end_pos=%d", regexEnd),
							err,
						),
					)
				} else {
					constraints = append(constraints, constraint)
				}
				mode = typeMode
				// Skip past any whitespace and comma to next constraint
				for pos < last && isWhitespace(spec[pos]) {
					pos++
				}
				if pos < last && spec[pos] == ',' {
					pos++
					for pos < last && isWhitespace(spec[pos]) {
						pos++
					}
					constraintStart = pos
				}
				continue
			}

			if ch == ']' {
				// Look ahead to see if this ends the constraint (comma or end of string)
				isEndOfConstraint := false
				if pos == last {
					// End of string
					isEndOfConstraint = true
				} else {
					// Check for optional whitespace followed by comma or end
					lookahead := pos
					for lookahead < last && isWhitespace(spec[lookahead]) {
						lookahead++
					}
					if lookahead == last || spec[lookahead] == ',' {
						isEndOfConstraint = true
					}
				}

				if isEndOfConstraint {
					// This closes the constraint
					value = spec[valueStart : pos-1]
					constraint, err = constraint.Parse(value, dataType)
					if err != nil {
						errs = append(errs,
							errors.Join(ErrParseFailed,
								fmt.Errorf("constraint_value=%s", value),
								fmt.Errorf("constraint_type=%s", ct),
								fmt.Errorf("constraint_spec=%s", spec),
								fmt.Errorf("data_type=%s", dataType.Slug()),
								fmt.Errorf("start_pos=%d", valueStart),
								fmt.Errorf("end_pos=%d", pos-1),
								err,
							),
						)
					} else {
						constraints = append(constraints, constraint)
					}
					mode = typeMode
					// Skip past any whitespace and comma to next constraint
					for pos < last && isWhitespace(spec[pos]) {
						pos++
					}
					if pos < last && spec[pos] == ',' {
						pos++
						for pos < last && isWhitespace(spec[pos]) {
							pos++
						}
						constraintStart = pos
					}
					continue
				}
			}
			if pos == last {
				// End of string without closing bracket - malformed
				errs = append(errs,
					errors.Join(ErrInvalidSyntax,
						fmt.Errorf("position=%d", pos),
						fmt.Errorf("constraint_type=%s", ct),
						fmt.Errorf("constraint_spec=%s", spec),
						fmt.Errorf("data_type=%s", dataType.Slug()),
						errors.New("constraint value not properly closed"),
					),
				)
				continue
			}
		}
	}

end:
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return constraints, err
}

// findRegexBoundaries uses bidirectional parsing to find the true boundaries of a regex constraint.
// Returns (-1, -1) if no regex constraint is found.
// This handles regex patterns that contain [ and ] characters by:
// 1. Finding "regex[" from the start
// 2. Finding the last "]" that could close the regex
// 3. If other constraints follow, finding the ] before them
func findRegexBoundaries(spec string) (start, end int) {
	// Find "regex[" - the start of regex constraint
	regexPrefix := "regex["
	start = strings.Index(spec, regexPrefix)
	if start == -1 {
		return -1, -1 // No regex constraint
	}

	// Find the last "]" in the spec - this is our candidate end
	end = strings.LastIndex(spec, "]")
	if end == -1 || end <= start+len(regexPrefix) {
		return -1, -1 // No closing bracket or it's before/at the opening
	}

	// Check if there are other constraints after regex
	// Look for a comma after the potential regex end
	afterEnd := end + 1
	if afterEnd < len(spec) {
		// Skip whitespace
		for afterEnd < len(spec) && isWhitespace(spec[afterEnd]) {
			afterEnd++
		}
		// If we find a comma, there might be another constraint
		// The last ] we found is correct
		if afterEnd < len(spec) && spec[afterEnd] == ',' {
			// Keep the end position - it's the last ] before the comma
		}
	}

	return start, end
}

// isConstraintTypeChar returns true if the character is valid in a constraint type name.
// Constraint type names can contain lowercase letters and underscores.
func isConstraintTypeChar(ch byte) (isChar bool) {
	if 'a' <= ch && ch <= 'z' {
		isChar = true
		goto end
	}
	if ch == '_' {
		isChar = true
		goto end
	}
end:
	return isChar
}

// isWhitespace returns true if the character is considered whitespace.
// Recognized whitespace characters include space, tab, newline, and carriage return.
func isWhitespace(ch byte) (isWS bool) {
	switch ch {
	case ' ', '\t', '\n', '\r':
		isWS = true
	}
	return isWS
}
