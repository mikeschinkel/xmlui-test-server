package apipkg

import (
	"errors"
	"fmt"
	"strings"
)

type RowsExpected string

const (
	OneRow         RowsExpected = "one"
	ManyRows       RowsExpected = "many"
	OneRowOrNone   RowsExpected = "one?"
	ManyRowsOrNone RowsExpected = "many?"
)

func ParseRowsExpected(s string) (re RowsExpected, err error) {
	if s == "" {
		re = DefaultRowsExpected
		goto end
	}
	re = RowsExpected(strings.ToLower(s))
	switch re {
	case ManyRows, OneRow, ManyRowsOrNone, OneRowOrNone:

		// Nothing to do
	default:
		err = errors.Join(ErrInvalidRowsExpectedType, fmt.Errorf("rows_expected=%s", s))
		re = ""
	}
end:
	return re, err
}
