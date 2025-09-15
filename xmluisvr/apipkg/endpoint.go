package apipkg

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type Endpoint struct {
	Endpoint     common.URLPath
	Description  string
	Query        string
	QueryFile    common.Filepath
	Params       Params
	RowsExpected common.Cardinality
	RowType      common.DataType
	ColumnTypes  []common.DataType // TODO make this a bespoke column type
}

func ParseEndpoints(cfgEPs []*cfgldr.APIEndpointV2) (eps []*Endpoint, err error) {
	var errs []error
	eps = make([]*Endpoint, len(cfgEPs))
	for i, cfgEP := range cfgEPs {
		eps[i], err = ParseEndpoint(cfgEP)
		errs = append(errs, err)
	}
	return eps, errors.Join(errs...)
}

func ParseEndpoint(cfgEP *cfgldr.APIEndpointV2) (ep *Endpoint, err error) {
	var errs []error

	ep = &Endpoint{
		Description: cfgEP.Description,
		Query:       cfgEP.Query,
	}
	ep.Endpoint, err = common.ParseURLPath(cfgEP.Endpoint)
	errs = append(errs, err)
	ep.QueryFile, err = common.ParseFilepath(cfgEP.QueryFile)
	errs = append(errs, err)
	ep.Params, err = ParseParams(cfgEP.Params)
	errs = append(errs, err)
	ep.RowsExpected, err = common.ParseCardinality(cfgEP.Cardinality)
	errs = append(errs, err)
	ep.RowType, err = common.ParseRowType(cfgEP.RowType)
	errs = append(errs, err)
	ep.ColumnTypes, err = common.ParseColumnTypes(cfgEP.ColumnTypes)
	errs = append(errs, err)
	if len(errs) != 0 {
		err = errors.Join(errs...)
	}
	if err != nil {
		ep = &Endpoint{}
	}

	return ep, err
}
