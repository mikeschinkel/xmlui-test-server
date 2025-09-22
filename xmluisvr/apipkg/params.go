package apipkg

import (
	"errors"
	"fmt"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

type Param struct {
	Name        common.Identifier
	Type        common.DBDataType
	Constraints []pathvars.Constraint
	RawValue    string
}

func ParseParams(cfgParams any) (params []Param, err error) {
	var errs []error
	// TODO Fix this after getting JSONv2 Loading fixed
	//apiParams := cfgParams.ApiParamsMap()
	//
	//params = make([]Param, 0, len(apiParams))
	//
	//for name, apiParam := range apiParams {
	//	var p Param
	//	p, err = ParseAPIParam(string(name), apiParam)
	//	if err != nil {
	//		errs = append(errs, err)
	//		continue
	//	}
	//	params = append(params, p)
	//}
	return params, errors.Join(errs...)
}

var ErrInvalidAPIEndpointParameter = errors.New("invalid API endpoint parameter")

func ParseAPIParam(name string, cfg cfgldr.APIParam) (p Param, err error) {
	var pn common.Identifier
	var dt common.DBDataType
	var cc []pathvars.Constraint
	var dtNum pathvars.PVDataType
	param, ok := cfg.(cfgldr.APIParamV1)
	if !ok {
		paramSpec, ok := cfg.(cfgldr.APIParamsMapValue)
		if !ok {
			err = errors.Join(ErrInvalidAPIEndpointParameter, fmt.Errorf("parameter=%v", cfg))
			goto end
		}
		param, err = cfgldr.ParseAPIParamV1(name, string(paramSpec))
	}
	if err != nil {
		goto end
	}

	pn, err = common.ParseIdentifier(param.Name)
	if err != nil {
		goto end
	}
	dt, err = common.ParseDBDataType(param.Type)
	if err != nil {
		goto end
	}
	dtNum, err = pathvars.ParsePVDataType(param.Type)
	if err != nil {
		goto end
	}
	cc, err = pathvars.ParseConstraints(param.Constraints, dtNum)
	if err != nil {
		goto end
	}
	p = Param{
		Name:        pn,
		Type:        dt,
		Constraints: cc,
	}
end:
	return p, err
}
