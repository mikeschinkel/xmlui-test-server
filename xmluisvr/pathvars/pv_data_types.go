// Package pathvars/pv_data_types defines the data types supported for path and query parameters.
// It provides type definitions, name mappings, and parsing functions for validating
// parameter values against specific data types like integers, UUIDs, dates, etc.
package pathvars

import (
	"errors"
	"fmt"
)

// init registers type aliases for commonly used alternate names.
func init() {
	RegisterDataTypeAlias(IntegerType, IntTypeName)
	RegisterDataTypeAlias(AlphanumericType, AlphanumTypeName)
	RegisterDataTypeAlias(BooleanType, BoolTypeName)
}

// Default parameter data type constants.
const (
	// DefaultPVDataType is the default data type used when no type is specified.
	DefaultPVDataType = StringType

	// DefaultPVDataTypeName is the string representation of the default data type.
	DefaultPVDataTypeName PVDataTypeSlug = StringTypeName
)

// PVDataType represents the enumerated data types supported for parameters.
type PVDataType int

// Supported parameter data types.
const (
	// UnspecifiedDataType indicates no data type was specified.
	UnspecifiedDataType PVDataType = iota

	// StringType represents text data with no specific format requirements.
	StringType

	// IntegerType represents whole number values (positive, negative, or zero).
	IntegerType

	// RealType represents floating-point numeric values.
	RealType

	// DecimalType represents decimal numeric values with precise fractional parts.
	DecimalType

	// IdentifierType represents programming-style identifiers (lowercase, alphanumeric with underscores).
	IdentifierType

	// DateType represents date/time values that can be validated against various formats.
	DateType

	// UUIDType represents Universally Unique Identifier values.
	UUIDType

	// AlphanumericType represents values containing only letters and digits.
	AlphanumericType

	// SlugType represents URL-safe slug values (lowercase, hyphen-separated).
	SlugType

	// BooleanType represents true/false values.
	BooleanType

	// EmailType represents email address values.
	EmailType
)

// PVDataTypeSlug represents the string name of a parameter data type.
type PVDataTypeSlug string

// String names for parameter data types.
const (
	// InvalidTypeName indicates an unrecognized type name.
	InvalidTypeName PVDataTypeSlug = "invalid"

	// StringTypeName is the string representation of StringType.
	StringTypeName PVDataTypeSlug = "string"

	// IntegerTypeName is the string representation of IntegerType.
	IntegerTypeName PVDataTypeSlug = "integer"

	// IntTypeName is an accepted alternate name for IntegerType.
	IntTypeName PVDataTypeSlug = "int" // Accepted alternate for "integer"

	// DecimalTypeName is the string representation of DecimalType.
	DecimalTypeName PVDataTypeSlug = "decimal"

	// RealTypeName is the string representation of RealType.
	RealTypeName PVDataTypeSlug = "real"

	// IdentifierTypeName is the string representation of IdentifierType.
	IdentifierTypeName PVDataTypeSlug = "identifier"

	// DateTypeName is the string representation of DateType.
	DateTypeName PVDataTypeSlug = "date"

	// UUIDTypeName is the string representation of UUIDType.
	UUIDTypeName PVDataTypeSlug = "uuid"

	// AlphanumericTypeName is the string representation of AlphanumericType.
	AlphanumericTypeName PVDataTypeSlug = "alphanumeric"

	// AlphanumTypeName is an accepted alternate name for AlphanumericType.
	AlphanumTypeName PVDataTypeSlug = "alphanum" // Accepted alternate for "alphanumeric"

	// SlugTypeName is the string representation of SlugType.
	SlugTypeName PVDataTypeSlug = "slug"

	// BooleanTypeName is the string representation of BooleanType.
	BooleanTypeName PVDataTypeSlug = "boolean"

	// BoolTypeName is an accepted alternate name for BooleanType.
	BoolTypeName PVDataTypeSlug = "bool" // Accepted alternate for "boolean"

	// EmailTypeName is the string representation of EmailType.
	EmailTypeName PVDataTypeSlug = "email"
)

func (dt PVDataType) WithIndefiniteArticle() string {
	slug := string(dt.Slug())
	switch dt {
	case AlphanumericType, IdentifierType, IntegerType:
		return "an " + slug
	default:
		// Stop Goland from complaining switch {} is incomplete in its case statements
	}
	return "a" + slug
}

// Slug returns the canonical lowercase string name for this data type	.
func (dt PVDataType) Slug() PVDataTypeSlug {
	switch dt {
	case StringType:
		return StringTypeName
	case IntegerType:
		return IntegerTypeName
	case DecimalType:
		return DecimalTypeName
	case RealType:
		return RealTypeName
	case IdentifierType:
		return IdentifierTypeName
	case DateType:
		return DateTypeName
	case UUIDType:
		return UUIDTypeName
	case AlphanumericType:
		return AlphanumericTypeName
	case SlugType:
		return SlugTypeName
	case BooleanType:
		return BooleanTypeName
	case EmailType:
		return EmailTypeName
	case UnspecifiedDataType:
		fallthrough
	default:
		return InvalidTypeName
	}
}

func (dt PVDataType) Example() any {
	switch dt {
	case StringType:
		return "abc"
	case IntegerType:
		return 123
	case DecimalType:
		// TODO Might need to consider format constraints
		return 1.23
	case RealType:
		return 1.2345
	case IdentifierType:
		return "id"
	case DateType:
		// TODO Might need to consider format constraints
		return "1999-12-31"
	case UUIDType:
		// TODO Might need to consider format constraints
		return "NEED A GOOD EXAMPLE"
	case AlphanumericType:
		return "abc123"
	case SlugType:
		return "abc-123"
	case BooleanType:
		return "true"
	case EmailType:
		return "yourname@example.com"
	case UnspecifiedDataType:
		fallthrough
	default:
		return "Unspecified has no example"
	}
}

// ParsePVDataType converts a string type name to a PVDataType enum value.
// Returns an error if the type name is not recognized.
func ParsePVDataType(typeStr string) (dataType PVDataType, err error) {
	switch PVDataTypeSlug(typeStr) {
	case StringTypeName:
		dataType = StringType
	case IntegerTypeName, IntTypeName:
		dataType = IntegerType
	case DecimalTypeName:
		dataType = DecimalType
	case RealTypeName:
		dataType = RealType
	case IdentifierTypeName:
		dataType = IdentifierType
	case DateTypeName:
		dataType = DateType
	case UUIDTypeName:
		dataType = UUIDType
	case AlphanumericTypeName, AlphanumTypeName:
		dataType = AlphanumericType
	case SlugTypeName:
		dataType = SlugType
	case BooleanTypeName, BoolTypeName:
		dataType = BooleanType
	case EmailTypeName:
		dataType = EmailType
	case InvalidTypeName:
		fallthrough
	default:
		err = errors.Join(
			ErrInvalidParameterType,
			ErrUnsupportedDataType,
			fmt.Errorf("date_type=%s", typeStr),
		)
	}
	return dataType, err
}

// GetDataType returns the data type if matched, or UnspecifiedDataType if not.
func GetDataType(name Identifier) PVDataType {
	dataType, err := ParsePVDataType(string(name))
	if err != nil {
		return UnspecifiedDataType
	}
	return dataType
}
