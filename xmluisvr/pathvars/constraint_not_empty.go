package pathvars

import (
	"fmt"
)

func init() {
	RegisterConstraint(&NotEmptyConstraint{})
}

var _ Constraint = (*NotEmptyConstraint)(nil)

// NotEmptyConstraint validates that a value is not empty
type NotEmptyConstraint struct {
	baseConstraint
}

func NewNotEmptyConstraint() *NotEmptyConstraint {
	c := &NotEmptyConstraint{}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *NotEmptyConstraint) ValidDataTypes() []PVDataType {
	return []PVDataType{
		AlphanumericType,
		DateType,
		DecimalType,
		EmailType,
		IdentifierType,
		IntegerType,
		RealType,
		SlugType,
		StringType,
		UUIDType,
	}
}

func (c *NotEmptyConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseNotEmptyConstraint(value)
}

func (c *NotEmptyConstraint) Type() ConstraintType {
	return NotEmptyConstraintType
}

func (c *NotEmptyConstraint) Validate(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func (c *NotEmptyConstraint) Rule() string {
	return ""
}

func (c *NotEmptyConstraint) String() string {
	return string(NotEmptyConstraintType)
}

func (c *NotEmptyConstraint) ErrorDetail(param *Parameter, value string) string {
	return fmt.Sprintf("Parameter '%s' with value '%s' failed constraint validation: value cannot be empty",
		param.Name,
		value,
	)
}

func (c *NotEmptyConstraint) ErrorSuggestion(param *Parameter, value, example string) string {
	return fmt.Sprintf("Ensure parameter '%s' satisfies the constraint: %s, for example: %s",
		param.Name,
		c.String(),
		example,
	)
}

func (c *NotEmptyConstraint) Example(err error) any {
	// Return "example" as a meaningful non-empty string example
	// TODO Return a better example. "this-examples" for a slug?
	// This should also be able to provide a different example depending on database type.
	// For example, a string is not a valud value for an 'integer' data type.
	return "example"
}

// ParseNotEmptyConstraint parses a notempty constraint (no arguments expected)
func ParseNotEmptyConstraint(value string) (constraint *NotEmptyConstraint, err error) {
	if value != "" {
		err = fmt.Errorf("non-empty constraint does not accept arguments, got: %q", value)
		goto end
	}
	constraint = NewNotEmptyConstraint()

end:
	return constraint, err
}
