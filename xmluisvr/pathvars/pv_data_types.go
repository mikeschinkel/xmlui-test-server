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
	DefaultPVDataTypeName PVDataTypeName = StringTypeName
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

// PVDataTypeName represents the string name of a parameter data type.
type PVDataTypeName string

// String names for parameter data types.
const (
	// InvalidTypeName indicates an unrecognized type name.
	InvalidTypeName PVDataTypeName = "invalid"

	// StringTypeName is the string representation of StringType.
	StringTypeName PVDataTypeName = "string"

	// IntegerTypeName is the string representation of IntegerType.
	IntegerTypeName PVDataTypeName = "integer"

	// IntTypeName is an accepted alternate name for IntegerType.
	IntTypeName PVDataTypeName = "int" // Accepted alternate for "integer"

	// DecimalTypeName is the string representation of DecimalType.
	DecimalTypeName PVDataTypeName = "decimal"

	// RealTypeName is the string representation of RealType.
	RealTypeName PVDataTypeName = "real"

	// IdentifierTypeName is the string representation of IdentifierType.
	IdentifierTypeName PVDataTypeName = "identifier"

	// DateTypeName is the string representation of DateType.
	DateTypeName PVDataTypeName = "date"

	// UUIDTypeName is the string representation of UUIDType.
	UUIDTypeName PVDataTypeName = "uuid"

	// AlphanumericTypeName is the string representation of AlphanumericType.
	AlphanumericTypeName PVDataTypeName = "alphanumeric"

	// AlphanumTypeName is an accepted alternate name for AlphanumericType.
	AlphanumTypeName PVDataTypeName = "alphanum" // Accepted alternate for "alphanumeric"

	// SlugTypeName is the string representation of SlugType.
	SlugTypeName PVDataTypeName = "slug"

	// BooleanTypeName is the string representation of BooleanType.
	BooleanTypeName PVDataTypeName = "boolean"

	// BoolTypeName is an accepted alternate name for BooleanType.
	BoolTypeName PVDataTypeName = "bool" // Accepted alternate for "boolean"

	// EmailTypeName is the string representation of EmailType.
	EmailTypeName PVDataTypeName = "email"
)

// TypeName returns the canonical string name for this data type.
func (dt PVDataType) TypeName() PVDataTypeName {
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

// ParsePVDataType converts a string type name to a PVDataType enum value.
// Returns an error if the type name is not recognized.
func ParsePVDataType(typeStr string) (dataType PVDataType, err error) {
	switch PVDataTypeName(typeStr) {
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
			fmt.Errorf("type=%q", typeStr),
			fmt.Errorf("reason=%s", "unsupported data type"),
		)
	}
	return dataType, err
}

// InferDataTypeFromName attempts to infer a data type from a parameter name.
// Returns the inferred data type and true if the name matches a known data type,
// otherwise returns UnspecifiedDataType and false.
func InferDataTypeFromName(name string) PVDataType {
	dataType, err := ParsePVDataType(name)
	if err != nil {
		return UnspecifiedDataType
	}
	return dataType
}
