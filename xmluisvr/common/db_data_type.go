package common

import (
	"errors"
	"fmt"
	"strings"
)

type DBDataType string

const (
	AnyDBDataType           DBDataType = "any"
	IntegerDBDataType       DBDataType = "integer"
	RealDBDataType          DBDataType = "real"
	StringDBDataType        DBDataType = "string"
	JSONDBDataType          DBDataType = "json"
	IntegerDBDataTypeOrNULL DBDataType = "integer?"
	RealDBDataTypeOrNULL    DBDataType = "real?"
	StringDBDataTypeOrNULL  DBDataType = "string?"
	JSONDBDataTypeOrNULL    DBDataType = "json?"
)

type DBRowType string

const (
	AnyRowType           DBRowType = "any"
	IntegerRowType       DBRowType = "integer"
	RealRowType          DBRowType = "real"
	StringRowType        DBRowType = "string"
	ColumnsRowType       DBRowType = "columns"
	JSONRowType          DBRowType = "json"
	IntegerRowTypeOrNULL DBRowType = "integer?"
	RealRowTypeOrNULL    DBRowType = "real?"
	StringRowTypeOrNULL  DBRowType = "string?"
	JSONRowTypeOrNULL    DBRowType = "json?"
)

func ParseDBRowType(s string) (rt DBRowType, err error) {
	if s == "" {
		rt = DefaultRowType
		goto end
	}
	rt = DBRowType(strings.ToLower(s))
	switch rt {
	case AnyRowType, IntegerRowType, RealRowType, StringRowType, ColumnsRowType, JSONRowType, IntegerRowTypeOrNULL, RealRowTypeOrNULL, StringRowTypeOrNULL, JSONRowTypeOrNULL:
		// Nothing to do
	default:
		err = errors.Join(ErrInvalidRowType, fmt.Errorf("row_type=%s", s))
		rt = ""
	}
end:
	return rt, err
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
		dt = DefaultDBDataType
		goto end
	}
	dt = DBDataType(strings.ToLower(s))
	switch dt {
	case AnyDBDataType, IntegerDBDataType, RealDBDataType, StringDBDataType, JSONDBDataType, IntegerDBDataTypeOrNULL, RealDBDataTypeOrNULL, StringDBDataTypeOrNULL, JSONDBDataTypeOrNULL:
		// Nothing to do
	default:
		err = errors.Join(ErrInvalidDataType, fmt.Errorf("data_type=%s", s))
		dt = ""
	}
end:
	return dt, err
}
