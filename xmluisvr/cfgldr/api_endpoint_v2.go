package cfgldr

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mikeschinkel/go-dt"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
	"github.com/xmlui-org/localsvr/xmluisvr/dbqvars"

	. "github.com/mikeschinkel/go-doterr"
)

// APIEndpointV2 is the main endpoint struct using JSONV2 inline to flatten the JSON
type APIEndpointV2 struct {
	APIEndpointBase `json:",inline"`
	Params          APIParamsMapper `json:"params"`
	paramsType      reflect.Type
}

// APIEndpointBase contains all the non-polymorphic properties for APIEndpointV2
type APIEndpointBase struct {
	Notes         []string `json:"@notes,omitempty"`
	Method        string   `json:"method"`
	Path          string   `json:"path"`
	Description   string   `json:"description"`
	Query         string   `json:"query"`
	QueryFile     string   `json:"query_file"`
	configDir     string
	queryFilepath string
	Cardinality   string   `json:"cardinality"`  // 'one' or 'many'
	RowType       string   `json:"row_type"`     // 'int', 'real','string','json','columns'
	ColumnTypes   []string `json:"column_types"` // used when row_type="columns"
}

func (ep APIEndpointBase) Endpoint() string {
	var method string
	var path string
	if ep.Method == "" {
		method = string(common.ANYMethod)
	} else {
		method = ep.Method
	}
	switch {
	case ep.Path == "":
		path = "/"
	case ep.Path[0] != '/':
		path = fmt.Sprintf("./%s", ep.Path)
	default:
		path = ep.Path
	}
	return fmt.Sprintf("%s %s", method, path)
}

type APIEndpointV2Args struct {
	Notes       []string
	Description string
	Query       string
	QueryFile   string
	Params      APIParamsMapper
	Cardinality string
	RowType     string
	ColumnTypes []string
	ParamsType  reflect.Type
}

func NewAPIEndpointV2(method, path string, args APIEndpointV2Args) *APIEndpointV2 {
	if args.Notes == nil {
		args.Notes = make([]string, 0)
	}
	if args.Params == nil {
		args.Params = APIParamsV1{}
	}
	if args.ColumnTypes == nil {
		args.ColumnTypes = make([]string, 0)
	}
	if args.ParamsType == nil {
		args.ParamsType = reflect.TypeOf(([]APIParamV1)(nil))
	}
	switch args.ParamsType {
	case reflect.TypeOf(([]APIParamV1)(nil)):
	case reflect.TypeOf((*APIParamsMap)(nil)):
	default:
		panic(fmt.Sprintf("Unsupported Parameters Type '%T' for endpoint %s %s'", args.ParamsType, method, path))
	}
	return &APIEndpointV2{
		APIEndpointBase: APIEndpointBase{
			Notes:       args.Notes,
			Method:      method,
			Path:        path,
			Description: args.Description,
			Query:       args.Query,
			QueryFile:   args.QueryFile,
			Cardinality: args.Cardinality,
			RowType:     args.RowType,
			ColumnTypes: args.ColumnTypes,
		},
		Params:     args.Params,
		paramsType: args.ParamsType,
	}
}

func (ep *APIEndpointV2) Normalize(sourceFile dt.Filepath, opts *Options) {
	if ep.Description == "" {
		ep.Description = ep.Endpoint()
	}
	if ep.Cardinality == "" {
		ep.Cardinality = string(dbqvars.DefaultCardinality)
	}
	if ep.RowType == "" {
		ep.RowType = string(dbqvars.DefaultRowType)
	}
	if ep.Params == nil {
		ep.Params = APIParamsV1{}
	}
	if ep.paramsType == nil {
		ep.paramsType = reflect.TypeOf(([]APIParamV1)(nil))
	}
	if ep.configDir == "" {
		ep.configDir = filepath.Dir(string(sourceFile))
	}
}

func (ep *APIEndpointV2) IsMapFormat() bool {
	return ep.paramsType == reflect.TypeOf((*APIParamsMap)(nil))
}

func (ep *APIEndpointV2) MarshalJSON() (json []byte, err error) {
	var apiParams APIParamsV1
	var ok bool

	// Create temporary struct for marshaling
	var temp struct {
		APIEndpointBase `json:",inline"`
		Params          any `json:"params"`
	}
	// Copy base fields
	temp.APIEndpointBase = ep.APIEndpointBase

	marshalFunc := func() ([]byte, error) {
		return jsonv2.Marshal(temp, jsontext.WithIndent("  "))
	}

	if !ep.IsMapFormat() {
		// Keep as array format
		temp.Params = ep.Params
		json, err = marshalFunc()
		goto end
	}

	// Convert params to original format based on paramsType
	apiParams, ok = ep.Params.(APIParamsV1)
	if !ok {
		err = NewErr(ErrAPIParamsIsAnInvalidDataType,
			"endpoint", ep.Endpoint(),
			fmt.Errorf("data_type=%T", ep.Params),
		)
		goto end
	}

	temp.Params = apiParams.APIParamsMap()
	json, err = marshalFunc()

end:
	return json, err
}

func (ep *APIEndpointV2) UnmarshalJSON(data []byte) (err error) {
	var isMap bool
	var errs []error

	// Create a temporary struct that matches RootConfigV1 but with DBConfig as RawMessage
	var temp struct {
		APIEndpointBase `json:",inline"`
		Params          jsontext.Value `json:"params"`
	}

	var params []APIParamV1
	var paramsMap APIParamsMap

	err = jsonv2.Unmarshal(data, &temp)
	if err != nil {
		goto end
	}

	ep.APIEndpointBase = temp.APIEndpointBase

	if []byte(temp.Params) == nil {
		ep.Params = APIParamsV1{}
		ep.paramsType = reflect.TypeOf(([]APIParamV1)(nil))
		goto end
	}

	err = jsonv2.Unmarshal(temp.Params, &params)
	if err != nil {
		isMap = true
		err = jsonv2.Unmarshal(temp.Params, &paramsMap)
	}
	if err != nil {
		goto end
	}
	if !isMap {
		ep.Params = APIParamsV1(params)
		ep.paramsType = reflect.TypeOf(([]APIParamV1)(nil))
		goto end
	}
	for name, spec := range paramsMap.Iterator() {
		if name != "" && name[0] == '@' {
			// Ignore comments
			continue
		}
		param, err := ParseAPIParamV1(string(name), string(spec))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		params = append(params, param)
	}
	ep.Params = APIParamsV1(params)
	ep.paramsType = reflect.TypeOf((*APIParamsMap)(nil))
	err = CombineErrs(errs)

end:
	return err
}

// ErrFailedToReadQueryFile indicates that an SQL file referenced by an endpoint could not be read.
var ErrFailedToReadQueryFile = errors.New("failed to read query file")
var ErrEitherQueryOrQueryFile = errors.New("both query file and query cannot have values")

func (ep *APIEndpointV2) GetQuery() (q string, err error) {
	var queryBytes []byte

	q = ep.Query

	if q != "" && ep.QueryFile != "" {
		err = NewErr(ErrEitherQueryOrQueryFile,
			"endpoint", ep.Endpoint(),
			"query", leftN(ep.QueryFile, 50),
			"query_file", ep.QueryFile,
		)
	}
	// Check if SQL should be loaded from a file
	if ep.QueryFile == "" {
		// Use the inline SQL from the APIConfig definition
		goto end
	}
	ep.queryFilepath = ep.GetQueryFilepath()
	if err != nil {
	}
	// Determine the APIConfig description file's directory to make relative paths work

	// Read the SQL file
	queryBytes, err = os.ReadFile(ep.queryFilepath)
	if err != nil {
		err = NewErr(ErrFailedToReadQueryFile,
			"endpoint", ep.Endpoint(),
			"query", leftN(ep.QueryFile, 50),
			"query_file", ep.QueryFile,
			err,
		)
		goto end
	}

	q = string(queryBytes)
end:
	return q, err
}

func (ep *APIEndpointV2) GetQueryFilepath() string {
	if ep.configDir == "" {
		panic(fmt.Sprintf("Config Directory not set for %s", ep.Endpoint()))
	}
	if ep.queryFilepath != "" {
		goto end
	}
	// Build the SQL file path relative to the APIConfig description file
	ep.queryFilepath = filepath.Join(ep.configDir, string(ep.QueryFile))
end:
	return ep.queryFilepath
}

func leftN(s string, n int) string {
	if len(s) < n {
		goto end
	}
	s = s[:n]
end:
	return s
}
