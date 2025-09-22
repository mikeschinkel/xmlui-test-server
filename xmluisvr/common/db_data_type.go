package common

import (
	"errors"
	"fmt"
	"strings"
)

type DBDataType string

const (
	AnyRowType           DBDataType = "any"
	IntegerRowType       DBDataType = "integer"
	RealRowType          DBDataType = "real"
	StringRowType        DBDataType = "string"
	ColumnsRowType       DBDataType = "columns"
	JSONRowType          DBDataType = "json"
	IntegerRowOrNULLType DBDataType = "integer?"
	RealRowOrNULLType    DBDataType = "real?"
	StringRowOrNULLType  DBDataType = "string?"
	JSONRowOrNULLType    DBDataType = "json?"
)

func ParseRowType(s string) (dt DBDataType, err error) {
	if s == "" {
		dt = DefaultRowType
		goto end
	}
	dt, err = ParseDBDataType(s)
	if err != nil {
		err = errors.Join(ErrInvalidRowType, err)
	}
end:
	return dt, err
}

func ParseColumnTypes(ss []string) (cts []DBDataType, err error) {
	var errs []error
	if len(ss) == 0 {
		goto end
	}
	cts = make([]DBDataType, len(ss))
	for i, s := range ss {
		value, err := ParseDBDataType(s)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cts[i] = value
	}
	err = errors.Join(errs...)
	if err != nil {
		err = errors.Join(ErrInvalidResultsColumnDataType, err)
	}
end:
	return cts, err
}

func ParseDBDataType(s string) (dt DBDataType, err error) {
	if s == "" {
		dt = DefaultDataType
		goto end
	}
	dt = DBDataType(strings.ToLower(s))
	switch dt {
	case AnyRowType, IntegerRowType, RealRowType, StringRowType, ColumnsRowType, JSONRowType, IntegerRowOrNULLType, RealRowOrNULLType, StringRowOrNULLType, JSONRowOrNULLType:
		// Nothing to do
	default:
		err = errors.Join(ErrInvalidDataType, fmt.Errorf("data_type=%s", s))
		dt = ""
	}
end:
	return dt, err
}
