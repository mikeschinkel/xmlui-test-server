package apipkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

func ParseEndpoints(cfgEPs []*cfgldr.APIEndpointV2) (eps []*Endpoint, err error) {
	var errs []error
	eps = make([]*Endpoint, len(cfgEPs))
	for i, cfgEP := range cfgEPs {
		eps[i], err = ParseEndpoint(cfgEP)
		errs = append(errs, err)
	}
	return eps, errors.Join(errs...)
}

var ErrOneOfQueryAndQueryFileMustNotBeEmpty = errors.New("at least one of query for query file must not be empty")

func ParseEndpoint(cfg *cfgldr.APIEndpointV2) (ep *Endpoint, err error) {
	var errs []error

	ep = &Endpoint{
		Description: cfg.Description,
		Query:       common.QueryString(cfg.Query),
		QueryFile:   common.Filepath(cfg.QueryFile),
	}
	method, path := splitEndPoint(cfg.Endpoint)
	// TODO: Should we allow defining endpoints without explicitly specifying a method?
	ep.method, err = common.ParseHTTPMethod(method, common.EmptyOk)
	errs = append(errs, err)
	ep.path, err = common.ParseURLPath(path)
	errs = append(errs, err)
	ep.QueryFile, err = common.ParseFilepath(cfg.QueryFile)
	errs = append(errs, err)
	ep.Params, err = ParseParams(cfg.Params)
	errs = append(errs, err)
	ep.RowsExpected, err = common.ParseCardinality(cfg.Cardinality)
	errs = append(errs, err)
	ep.RowType, err = common.ParseDBRowType(cfg.RowType)
	errs = append(errs, err)
	ep.ColumnTypes, err = common.ParseColumnTypes(cfg.ColumnTypes)
	errs = append(errs, err)
	if ep.Query == "" && ep.QueryFile == "" {
		errs = append(errs, ErrOneOfQueryAndQueryFileMustNotBeEmpty)
	}
	err = errors.Join(errs...)
	if err != nil {
		*ep = Endpoint{}
		err = errors.Join(err, fmt.Errorf("endpoint=%s", cfg.Endpoint))
	}

	return ep, err
}

type EndPointString string

type Endpoint struct {
	Description   string
	Query         common.QueryString
	QueryFile     common.Filepath
	queryFilepath common.Filepath
	Params        []Param
	RowsExpected  common.Cardinality
	RowType       common.DBRowType
	ColumnTypes   []common.DBDataType
	method        common.HTTPMethod
	path          common.URLPath
}

func (ep *Endpoint) PathVarsParameters() (params []pathvars.Parameter) {
	params = make([]pathvars.Parameter, 0, len(ep.Params))
	for _, p := range ep.Params {
		panic("FINISH THIS")
		param := pathvars.NewParameter(pathvars.ParameterArgs{
			Name:         string(p.Name),
			ParamType:    pathvars.UnspecifiedParameterType,
			DataType:     0,
			Constraints:  nil,
			Position:     0,
			Original:     "",
			MultiSegment: false,
			Optional:     false,
			DefaultValue: nil,
		})
		params = append(params, param)
	}
	return params
}

func (ep *Endpoint) GetQuery(dir common.DirPath) (q common.QueryString, queryFile common.Filepath, err error) {
	var queryBytes []byte

	q = ep.Query

	// Check if SQL should be loaded from a file
	if ep.QueryFile == "" {
		// Use the inline SQL from the APIConfig definition
		goto end
	}
	if ep.queryFilepath != "" {
		goto end
	}
	// Determine the APIConfig description file's directory to make relative paths work

	// Build the SQL file path relative to the APIConfig description file
	ep.queryFilepath = common.Filepath(filepath.Join(string(dir), string(ep.QueryFile)))

	// Read the SQL file
	queryBytes, err = os.ReadFile(string(queryFile))
	if err != nil {
		err = errors.Join(ErrFailedToReadQueryFile, err)
		goto end
	}

	q = common.QueryString(queryBytes)
end:
	return q, ep.queryFilepath, err
}

func (ep *Endpoint) Endpoint() EndPointString {
	return EndPointString(fmt.Sprintf("%s %s",
		strings.ToUpper(string(ep.method)),
		ep.path),
	)
}

func (ep *Endpoint) Path() common.URLPath {
	return ep.path
}

func (ep *Endpoint) Method() (m common.HTTPMethod) {
	m = ep.RawMethod()
	if m == common.ANYMethod {
		m = ""
		goto end
	}
end:
	return m
}
func (ep *Endpoint) RawMethod() common.HTTPMethod {
	return ep.method
}

func splitEndPoint(ep string) (method, path string) {
	method, path, found := strings.Cut(ep, " ")
	if !found {
		path = string(ep)
		goto end
	}
end:
	return method, path
}
