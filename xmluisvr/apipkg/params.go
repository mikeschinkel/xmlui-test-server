package apipkg

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type Params map[common.Identifier]common.DataType

func ParseParams(cfgParams map[string]string) (params Params, err error) {
	var errs []error
	params = make(Params)
	for name, typ := range cfgParams {
		var id common.Identifier
		id, err = common.ParseIdentifier(name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		dt, err := common.ParseDataType(typ)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		params[id] = dt
	}
	return params, errors.Join(errs...)
}
