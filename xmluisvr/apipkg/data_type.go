package apipkg

import (
	"errors"
	"fmt"
	"strings"
)

type DataType string

const (
	AnyRowType           DataType = "any"
	IntegerRowType       DataType = "integer"
	RealRowType          DataType = "real"
	StringRowType        DataType = "string"
	ColumnsRowType       DataType = "columns"
	JSONRowType          DataType = "json"
	IntegerRowOrNULLType DataType = "integer?"
	RealRowOrNULLType    DataType = "real?"
	StringRowOrNULLType  DataType = "string?"
	JSONRowOrNULLType    DataType = "json?"
)

func ParseRowType(s string) (dt DataType, err error) {
	if s == "" {
		dt = DefaultRowType
		goto end
	}
	dt, err = ParseDataType(s)
	if err != nil {
		err = errors.Join(ErrInvalidRowType, err)
	}
end:
	return dt, err
}

func ParseColumnTypes(ss []string) (cts []DataType, err error) {
	var errs []error
	cts = make([]DataType, len(ss))
	for i, s := range ss {
		cts[i], err = ParseDataType(s)
		errs = append(errs)
	}
	err = errors.Join(errs...)
	if err != nil {
		err = errors.Join(ErrInvalidResultsColumnDataType, err)
	}
	return cts, err
}

func ParseDataType(s string) (dt DataType, err error) {
	if s == "" {
		dt = DefaultDataType
		goto end
	}
	dt = DataType(strings.ToLower(s))
	switch dt {
	case IntegerRowType, RealRowType, StringRowType, ColumnsRowType, JSONRowType, IntegerRowOrNULLType, RealRowOrNULLType, StringRowOrNULLType, JSONRowOrNULLType:
		// Nothing to do
	default:
		err = errors.Join(ErrInvalidDataType, fmt.Errorf("data_type=%s", s))
		dt = ""
	}
end:
	return dt, err
}
