package dbpkg

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"

	. "github.com/xmlui-org/xmlui-test-server/xmluisvr/doterr"
)

type (
	ColumnName string
	TableCell  any
	TableRow   map[ColumnName]TableCell
	TableRows  []TableRow
)

type QueryResult TableRows

func (qr QueryResult) GetByCardinality(c dbqvars.Cardinality) (content any, err error) {
	switch c {
	case dbqvars.ManyRowsOrNone:
		if len(qr) == 0 {
			goto end
		}

	case dbqvars.ManyRows:
		if len(qr) == 0 {
			err = ErrManyRowsExpectedZeroReturned
			goto end
		}

	case dbqvars.OneRowOrNone:
		if len(qr) == 0 {
			goto end
		}
		fallthrough

	case dbqvars.OneRow:
		if len(qr) == 0 {
			err = ErrOneRowExpectedZeroReturned
			goto end
		}
		if len(qr) > 1 {
			err = NewErr(
				ErrOneRowExpectedManyReturned,
				"row_count", len(qr),
			)
			goto end
		}
		content = TableRows(qr)[0]

	default:
		content = qr
	}

end:
	if err != nil {
		err = WithErr(ErrInvalidCardinality, err)
	}
	return qr, err
}
