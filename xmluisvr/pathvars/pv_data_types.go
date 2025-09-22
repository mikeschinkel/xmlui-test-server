package pathvars

import (
	"errors"
	"fmt"
)

func init() {
	RegisterDataTypeAlias(IntegerType, IntTypeName)
	RegisterDataTypeAlias(AlphanumericType, AlphanumTypeName)
	RegisterDataTypeAlias(BooleanType, BoolTypeName)
}

const (
	DefaultPVDataType                    = StringType
	DefaultPVDataTypeName PVDataTypeName = StringTypeName
)

// PVDataType represents the type of a parameter
type PVDataType int

const (
	UnspecifiedType PVDataType = iota
	StringType
	IntegerType
	RealType
	DecimalType
	IdentifierType
	DateType
	UUIDType
	AlphanumericType
	SlugType
	BooleanType
	EmailType
)

type PVDataTypeName string

const (
	InvalidTypeName      PVDataTypeName = "invalid"
	StringTypeName       PVDataTypeName = "string"
	IntegerTypeName      PVDataTypeName = "integer"
	IntTypeName          PVDataTypeName = "int" // Accepted alternate for "integer"
	DecimalTypeName      PVDataTypeName = "decimal"
	RealTypeName         PVDataTypeName = "real"
	IdentifierTypeName   PVDataTypeName = "identifier"
	DateTypeName         PVDataTypeName = "date"
	UUIDTypeName         PVDataTypeName = "uuid"
	AlphanumericTypeName PVDataTypeName = "alphanumeric"
	AlphanumTypeName     PVDataTypeName = "alphanum" // Accepted alternate for "alphanumeric"
	SlugTypeName         PVDataTypeName = "slug"
	BooleanTypeName      PVDataTypeName = "boolean"
	BoolTypeName         PVDataTypeName = "bool" // Accepted alternate for "boolean"
	EmailTypeName        PVDataTypeName = "email"
)

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
	case UnspecifiedType:
		fallthrough
	default:
		return InvalidTypeName
	}
}

// ParsePVDataType converts string type to PVDataType enum
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
			ErrInvalidType,
			fmt.Errorf("type=%q", typeStr),
			fmt.Errorf("reason=%s", "unsupported data type"),
		)
	}
	return dataType, err
}

// InferDataTypeFromName attempts to infer a data type from a parameter name
// Returns the inferred data type and true if the name matches a data type, otherwise UnspecifiedType and false
func InferDataTypeFromName(name string) (PVDataType, bool) {
	dataType, err := ParsePVDataType(name)
	if err != nil {
		return UnspecifiedType, false
	}
	return dataType, true
}
