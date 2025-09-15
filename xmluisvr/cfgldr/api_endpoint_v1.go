package cfgldr

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type APIEndpointV2 struct {
	Endpoint    string            `json:"endpoint"`
	Description string            `json:"description"`
	Query       string            `json:"query"`
	QueryFile   string            `json:"query_file"`
	Params      map[string]string `json:"params"`
	Cardinality string            `json:"cardinality"`  // 'one' or 'many'
	RowType     string            `json:"row_type"`     // 'int','real','string','json','columns'
	ColumnTypes []string          `json:"column_types"` // used when row_type="columns"
	Method      string            `json:"-"`
	Path        string            `json:"-"`
}

var httpMethodsForRegexp = strings.Join(common.HTTPMethods, "|")

var endpointRegexp = regexp.MustCompile(`^(` + httpMethodsForRegexp + `)\s*(.*)$`)

func (ep *APIEndpointV2) Normalize() (err error) {
	var errs []error

	chkPath := true
	chkMethod := true

	if ep.Endpoint == "" {
		errs = append(errs, ErrAPIEndpointMustNotBeEmpty)
	} else {
		matches := endpointRegexp.FindAllStringSubmatch(ep.Endpoint, 2)
		switch {
		case matches != nil:
			ep.Method = matches[0][1]
			ep.Path = matches[0][2]
			chkMethod = false
		default:
			errs = append(errs, ErrInvalidAPIEndpoint)
			chkPath = false
		}
	}
	if chkMethod && ep.Endpoint != "" && ep.Method == "" {
		errs = append(errs, ErrInvalidAPIEndpointMethod)
	}
	if chkPath {
		_, err = common.ParseURLPath(ep.Path)
		errs = append(errs, ErrInvalidAPIEndpointPath)
	}
	if ep.Description == "" {
		ep.Description = ep.Endpoint
	}
	if ep.Query == "" && ep.QueryFile == "" {
		errs = append(errs, ErrAPIEndpointHasNoQueryOrFile)
	}
	if ep.Cardinality == "" {
		ep.Cardinality = string(common.DefaultCardinality)
	}
	var re common.Cardinality
	re, err = common.ParseCardinality(ep.Cardinality)
	if err != nil {
		errs = append(errs, err)
	} else {
		ep.Cardinality = string(re)
	}
	if ep.RowType == "" {
		ep.RowType = string(common.DefaultRowType)
	}
	var rt common.DataType
	rt, err = common.ParseRowType(ep.RowType)
	if err != nil {
		errs = append(errs, err)
	} else {
		ep.RowType = string(rt)
	}
	if ep.ColumnTypes == nil {
		ep.ColumnTypes = make([]string, 0)
	}
	errs = append(errs, ep.NormalizeParams())
	errs = append(errs, ep.NormalizeColumnTypes())

	if len(errs) > 0 {
		err = errors.Join(append(errs, fmt.Errorf("endpoint=%s", ep.Endpoint))...)
	}
	return err
}

func (ep *APIEndpointV2) NormalizeColumnTypes() (err error) {
	var errs []error
	if ep.ColumnTypes == nil {
		ep.ColumnTypes = make([]string, 0)
	}
	if len(ep.ColumnTypes) == 0 {
		goto end
	}
	for i, ct := range ep.ColumnTypes {
		value, err := common.ParseDataType(ct)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ep.ColumnTypes[i] = string(value)
	}
end:
	return errors.Join(errs...)
}

func (ep *APIEndpointV2) NormalizeParams() (err error) {
	var errs []error
	if ep.Params == nil {
		ep.Params = make(map[string]string)
	}
	if len(ep.Params) == 0 {
		goto end
	}
	for k, v := range ep.Params {
		id, err := common.ParseIdentifier(k)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// TODO Do values need to be validated?
		ep.Params[string(id)] = v
	}
end:
	return errors.Join(errs...)
}

type APIEndpointV2Args struct {
	Description string
	Query       string
	QueryFile   string
	Params      map[string]string
	Cardinality string
	RowType     string
	ColumnTypes []string
}

func NewAPIEndpointV2(endpoint string, args APIEndpointV2Args) *APIEndpointV2 {
	if args.Params == nil {
		args.Params = make(map[string]string)
	}
	if args.ColumnTypes == nil {
		args.ColumnTypes = make([]string, 0)
	}
	return &APIEndpointV2{
		Endpoint:    endpoint,
		Description: args.Description,
		Query:       args.Query,
		QueryFile:   args.QueryFile,
		Params:      args.Params,
		Cardinality: args.Cardinality,
		RowType:     args.RowType,
		ColumnTypes: args.ColumnTypes,
	}
}
