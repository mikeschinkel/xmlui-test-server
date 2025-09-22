package cfgldr

import (
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

var _ APIParam = (*APIParamV1)(nil)

type APIParamV1 struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Constraints string `json:"constraints"`
}

func (APIParamV1) APIParam() {}

func NewAPIParamV1(name string, typ string) APIParamV1 {
	return NewAPIParamV1WithConstraints(name, typ, "")
}

func NewAPIParamV1WithConstraints(name, typ, constraints string) APIParamV1 {
	return APIParamV1{
		Name:        name,
		Type:        typ,
		Constraints: constraints,
	}
}

func (p APIParamV1) String() string {
	sb := strings.Builder{}
	sb.WriteByte('{')
	sb.WriteString(strings.ToLower(p.Name))

	switch {
	case p.Type != "":
		sb.WriteByte(':')
		sb.WriteString(strings.ToLower(p.Type))
	case len(p.Constraints) != 0:
		sb.WriteByte(':')
	}

	if len(p.Constraints) != 0 {
		sb.WriteByte(':')
		sb.WriteString(p.Constraints)
	}
	sb.WriteByte('}')
	return sb.String()
}

// ParseAPIParamV1 parses the name and type/constraint spec and validates the
// type, returning an instance of APIParamV1 is a valid type, or an error
// otherwise.
func ParseAPIParamV1(name, spec string) (param APIParamV1, err error) {
	typ, cs, _ := strings.Cut(spec, ":")
	_, err = pathvars.ParsePVDataType(typ)
	if err != nil {
		// Type is explicitly specified, and invalid
		goto end
	}
	param = NewAPIParamV1WithConstraints(name, typ, cs)
end:
	return param, err
}
