package dbqvars

import (
	"errors"
	"fmt"
	"strings"
)

type Cardinality string

func (c Cardinality) EmptyOk() bool {
	return c[len(c)-1] == '?'
}

const (
	OneRow         Cardinality = "one"
	ManyRows       Cardinality = "many"
	OneRowOrNone   Cardinality = "one?"
	ManyRowsOrNone Cardinality = "many?"
)

func ParseCardinality(s string) (re Cardinality, err error) {
	if s == "" {
		re = DefaultCardinality
		goto end
	}
	re = Cardinality(strings.ToLower(s))
	switch re {
	case ManyRows, OneRow, ManyRowsOrNone, OneRowOrNone:

		// Nothing to do
	default:
		err = errors.Join(ErrInvalidCardinalityType, fmt.Errorf("cardinality=%s", s))
		re = ""
	}
end:
	return re, err
}
