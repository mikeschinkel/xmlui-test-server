package pathvars

import (
	"errors"
	"fmt"
	"strings"
)

func init() {
	RegisterConstraint(&EnumConstraint{})
}

var _ Constraint = (*EnumConstraint)(nil)

// EnumConstraint validates against allowed values
type EnumConstraint struct {
	baseConstraint
	values map[string]bool
	list   []string
}

func NewEnumConstraint(values map[string]bool, list []string) *EnumConstraint {
	c := &EnumConstraint{values: values, list: list}
	c.baseConstraint = newBaseConstraint(c)
	return c
}

func (c *EnumConstraint) Parse(value string, dataType PVDataType) (Constraint, error) {
	return ParseEnumConstraint(value)
}

func (c *EnumConstraint) Type() ConstraintType {
	return EnumConstraintType
}

func (c *EnumConstraint) Validate(value string) (err error) {
	if !c.values[value] {
		err = fmt.Errorf("value must be one of: %s", strings.Join(c.list, ", "))
	}
	return err
}

func (c *EnumConstraint) String() string {
	return fmt.Sprintf("%s[%s]", c.Type(), strings.Join(c.list, ","))
}

func (c *EnumConstraint) ValidDateTypes() []PVDataType {
	return []PVDataType{
		IntegerType,
		BooleanType,
		StringType,
		IdentifierType,
		AlphanumericType,
		SlugType,
		EmailType,
	}
}

// ParseEnumConstraint parses val1,val2,val3 format
func ParseEnumConstraint(enumSpec string) (constraint *EnumConstraint, err error) {
	var values []string
	var valueMap map[string]bool
	var value string

	if enumSpec == "" {
		err = errors.Join(
			ErrInvalidConstraint,
			fmt.Errorf("enumSpec=%q", enumSpec),
			fmt.Errorf("reason=%s", "empty enum content"),
		)
		goto end
	}

	// Split by comma
	values = strings.Split(enumSpec, ",")
	valueMap = make(map[string]bool)

	for _, value = range values {
		value = strings.TrimSpace(value)
		if value == "" {
			err = errors.Join(
				ErrInvalidConstraint,
				fmt.Errorf("enumSpec=%q", enumSpec),
				fmt.Errorf("reason=%s", "empty value in enum list"),
			)
			goto end
		}
		valueMap[value] = true
	}

	constraint = NewEnumConstraint(valueMap, values)

end:
	return constraint, err
}
