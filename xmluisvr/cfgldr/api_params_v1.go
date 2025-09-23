package cfgldr

import (
	"fmt"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

var _ APIParamsMapper = (*APIParamsV1)(nil)

type APIParamsV1 []APIParamV1

func (ps APIParamsV1) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (ps APIParamsV1) APIParamsMap() (pm *APIParamsMap) {
	pm = &APIParamsMap{
		OrderedMap: *NewOrderedMap[APIParamsMapKey, APIParamsMapValue](),
	}
	for _, p := range ps {
		var value string
		switch {
		case p.Type == "" && p.Constraints == "":
			// Just use default type
			typ := pathvars.DefaultPVDataTypeName
			dt, err := pathvars.ParsePVDataType(p.Name)
			if err == nil {
				typ = dt.TypeName()
			}
			value = string(typ)
		case p.Type != "" && p.Constraints == "":
			// Just type
			value = p.Type
		case p.Type == "" && p.Constraints != "":
			// Default type with constraints
			typ := pathvars.DefaultPVDataTypeName
			dt, err := pathvars.ParsePVDataType(p.Name)
			if err == nil {
				typ = dt.TypeName()
			}
			value = fmt.Sprintf("%s:%s", typ, p.Constraints)
		case p.Type != "" && p.Constraints != "":
			// Type with constraints
			value = fmt.Sprintf("%s:%s", p.Type, p.Constraints)
		}
		if value != "" {
			pm.Set(APIParamsMapKey(p.Name), APIParamsMapValue(value))
		}
	}
	return pm
}
