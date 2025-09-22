package cfgldr

import (
	"fmt"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

var _ APIParamsMapper = (*APIParamsV1)(nil)

type APIParamsV1 []APIParamV1

func (ps APIParamsV1) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (ps APIParamsV1) APIParamsMap() (pm *APIParamsMap) {
	pm = &APIParamsMap{}
	for _, p := range ps {
		sb := strings.Builder{}
		sb.WriteString(p.Name)
		details := fmt.Sprintf("%s:%s", p.Type, p.Constraints)
		switch {
		case len(details) == 1:
			continue
		case p.Type == "" && p.Constraints != "":
			sb.WriteByte(':')
			typ := pathvars.DefaultPVDataTypeName
			dt, err := pathvars.ParsePVDataType(p.Name)
			if err == nil {
				typ = dt.TypeName()
			}
			sb.WriteString(string(typ))
			sb.WriteByte(':')
			sb.WriteString(p.Constraints)
			continue
		case p.Type != "" && p.Constraints == "":
			sb.WriteByte(':')
			sb.WriteString(p.Type)
		case p.Type != "" && p.Constraints != "":
			sb.WriteByte(':')
			sb.WriteString(p.Type)
			sb.WriteByte(':')
			sb.WriteString(p.Constraints)
		}
		pm.Set(APIParamsMapKey(p.Name), APIParamsMapValue(sb.String()))
	}
	return pm
}
