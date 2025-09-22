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

func (c *NotEmptyConstraint) ValidDateTypes() []PVDataType {
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

func (c *NotEmptyConstraint) String() string {
	return string(c.Type())
}

// ParseNotEmptyConstraint parses a not-empty constraint (no arguments expected)
func ParseNotEmptyConstraint(value string) (constraint *NotEmptyConstraint, err error) {
	if value != "" {
		err = fmt.Errorf("non-empty constraint does not accept arguments, got: %q", value)
		goto end
	}
	constraint = NewNotEmptyConstraint()

end:
	return constraint, err
}
