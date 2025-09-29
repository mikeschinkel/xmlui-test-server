package dbqvars

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

var _ ParsedQuery = (*ParsedSQL)(nil)

type ParsedSQL struct {
	SQL        common.SQLQuery
	parameters []Parameter // ordered by first appearance, deduped by Name
}

func NewParsedSQL(SQL common.SQLQuery, parameters []Parameter) ParsedSQL {
	if parameters == nil {
		parameters = make([]Parameter, 0)
	}
	return ParsedSQL{
		SQL:        SQL,
		parameters: parameters,
	}
}

//func (ps ParsedSQL) GetValues(paramsMap map[common.Identifier]any, bodyJSON []byte) (values []any, err error) {
//	return QueryTokens(ps.Parameters).GetValues(paramsMap, bodyJSON)
//}

func (ps ParsedSQL) QueryString() common.QueryString {
	return common.QueryString(ps.SQL)
}

func (ps ParsedSQL) Parameters() (names Parameters) {
	return ps.parameters
}
