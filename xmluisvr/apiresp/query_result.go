package apiresp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

var _ ResponsePayload = (*QueryResult)(nil)

type QueryResult dbpkg.QueryResult

func (qr QueryResult) GetByCardinality(c dbqvars.Cardinality) (content any, err error) {
	switch c {
	case dbqvars.ManyRowsOrNone:
		if len(qr) == 0 {
			goto end
		}

	case dbqvars.ManyRows:
		if len(qr) == 0 {
			err = dbpkg.ErrManyRowsExpectedZeroReturned
			goto end
		}

	case dbqvars.OneRowOrNone:
		if len(qr) == 0 {
			goto end
		}
		fallthrough

	case dbqvars.OneRow:
		if len(qr) == 0 {
			err = dbpkg.ErrOneRowExpectedZeroReturned
			goto end
		}
		if len(qr) > 1 {
			err = errors.Join(
				dbpkg.ErrOneRowExpectedManyReturned,
				fmt.Errorf("row_count=%d", len(qr)),
			)
			goto end
		}
		content = dbpkg.TableRows(qr)[0]

	default:
		content = qr
	}

end:
	if err != nil {
		err = errors.Join(dbpkg.ErrInvalidCardinality, err)
	}
	return qr, err
}

func NewQueryResult(qr dbpkg.QueryResult) QueryResult {
	return QueryResult(qr)
}

func (QueryResult) ResponsePayload() {}
func (qr QueryResult) MIMEType() rfc9457.MIMEType {
	return rfc9457.ApplicationJSON
}
func (qr QueryResult) Error() string {
	return fmt.Sprintf("QueryResult\nrow_count=%d", len(qr))
}

func (qr QueryResult) HTTPStatusCode() int {
	if len(qr) == 0 {
		return http.StatusNotFound
	}
	return http.StatusOK
}
