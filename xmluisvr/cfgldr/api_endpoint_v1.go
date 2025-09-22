package cfgldr

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"reflect"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// APIEndpointV2 is the main endpoint struct using JSONV2 inline to flatten the JSON
type APIEndpointV2 struct {
	apiEndpointBase `json:",inline"`
	Params          APIParamsMapper `json:"params"`
	paramsType      reflect.Type
}

// APIEndpointV2 is the main endpoint struct using JSONV2 inline to flatten the JSON
type apiEndpointBase struct {
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
}

func NewAPIEndpointV2(endpoint string, args APIEndpointV2Args) *APIEndpointV2 {
	if args.Params == nil {
		args.Params = APIParamsV1{}
	}
	if args.ColumnTypes == nil {
		args.ColumnTypes = make([]string, 0)
	}
	return &APIEndpointV2{
		apiEndpointBase: apiEndpointBase{
			Endpoint:    endpoint,
			Description: args.Description,
			Query:       args.Query,
			QueryFile:   args.QueryFile,
			Cardinality: args.Cardinality,
			RowType:     args.RowType,
			ColumnTypes: args.ColumnTypes,
		},
		Params:     args.Params,
		paramsType: reflect.TypeOf(([]APIParamV1)(nil)),
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
}

func (ep *APIEndpointV2) UnmarshalJSON(data []byte) (err error) {
	var isMap bool
	var errs []error

	// Create a temporary struct that matches RootConfigV1 but with DBConfig as RawMessage
	var temp struct {
		apiEndpointBase `json:",inline"`
		Params          jsontext.Value `json:"params"`
	}

	var params []APIParamV1
	var paramsMap APIParamsMap

	err = jsonv2.Unmarshal(data, &temp)
	if err != nil {
		goto end
	}

	ep.apiEndpointBase = temp.apiEndpointBase

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

//func (ep *APIEndpointV2) UnmarshalJSON(data []byte) (err error) {
//	var isMap bool
//	// Create a temporary struct that matches RootConfigV1 but with DBConfig as RawMessage
//	var temp struct {
//		Endpoint    string          `json:"endpoint"`
//		Description string          `json:"description"`
//		Query       string          `json:"query"`
//		QueryFile   string          `json:"query_file"`
//		Params      json.RawMessage `json:"params"`
//		Cardinality string          `json:"cardinality"`  // 'one' or 'many'
//		RowType     string          `json:"row_type"`     // 'int', 'real','string','json','columns'
//		ColumnTypes []string        `json:"column_types"` // used when row_type="columns"
//	}
//
//	var params []APIParamV1
//	var paramsMapInfo struct {
//		Params apiParams `json:"params"`
//	}
//
//	err = jsonv2.Unmarshal(data, &temp)
//	if err != nil {
//		goto end
//	}
//
//	ep.Endpoint = temp.Endpoint
//	ep.Description = temp.Description
//	ep.Query = temp.Query
//	ep.QueryFile = temp.QueryFile
//	ep.Cardinality = temp.Cardinality
//	ep.RowType = temp.RowType
//	ep.ColumnTypes = temp.ColumnTypes
//
//	if temp.Params == nil {
//		ep.Params = APIParamsV1{}
//		goto end
//	}
//
//	err = jsonv2.Unmarshal(temp.Params, &params)
//	if err != nil {
//		isMap = true
//		err = jsonv2.Unmarshal(temp.Params, &paramsMapInfo)
//	}
//	if err != nil {
//		goto end
//	}
//	if !isMap {
//		ep.Params = APIParamsV1(params)
//		goto end
//	}
//
//	for name, details := range paramsMapInfo.Params {
//		typ, cs, found := strings.Cut(string(details), ":")
//		if !found {
//			typ = string(details)
//		}
//		params = append(params,
//			NewAPIParamV1WithConstraints(string(name), typ, cs),
//		)
//	}
//	ep.Params = APIParamsV1(params)
//
//end:
//	return err
//}
