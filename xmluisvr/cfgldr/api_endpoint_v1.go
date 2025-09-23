package cfgldr

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"reflect"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// APIEndpointV2 is the main endpoint struct using JSONV2 inline to flatten the JSON
type APIEndpointV2 struct {
	APIEndpointBase `json:",inline"`
	Params          APIParamsMapper `json:"params"`
	paramsType      reflect.Type
}

// APIEndpointBase contains all the non-polymorphic properties for APIEndpointV2
type APIEndpointBase struct {
	Endpoint    string   `json:"endpoint"`
	Description string   `json:"description"`
	Query       string   `json:"query"`
	QueryFile   string   `json:"query_file"`
	Cardinality string   `json:"cardinality"`  // 'one' or 'many'
	RowType     string   `json:"row_type"`     // 'int', 'real','string','json','columns'
	ColumnTypes []string `json:"column_types"` // used when row_type="columns"
}

type APIEndpointV2Args struct {
	Description string
	Query       string
	QueryFile   string
	Params      APIParamsMapper
	Cardinality string
	RowType     string
	ColumnTypes []string
	ParamsType  reflect.Type
}

func NewAPIEndpointV2(endpoint string, args APIEndpointV2Args) *APIEndpointV2 {
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
		panic(fmt.Sprintf("Unsupported Params Type '%T' for endpoint %s'", args.ParamsType, endpoint))
	}
	return &APIEndpointV2{
		APIEndpointBase: APIEndpointBase{
			Endpoint:    endpoint,
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

func (ep *APIEndpointV2) Normalize() {
	if ep.Description == "" {
		ep.Description = ep.Endpoint
	}
	if ep.Cardinality == "" {
		ep.Cardinality = string(common.DefaultCardinality)
	}
	if ep.RowType == "" {
		ep.RowType = string(common.DefaultRowType)
	}
	if ep.Params == nil {
		ep.Params = APIParamsV1{}
	}
	if ep.paramsType == nil {
		ep.paramsType = reflect.TypeOf(([]APIParamV1)(nil))
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
		err = errors.Join(ErrAPIParamsIsAnInvalidDataType,
			fmt.Errorf("endpoint=%s", ep.Endpoint),
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
	err = errors.Join(errs...)

end:
	return err
}
