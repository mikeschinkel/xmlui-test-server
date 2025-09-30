package dbqvars

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type ParsedQuery interface {
	QueryString() common.QueryString
	Parameters() Parameters
}
