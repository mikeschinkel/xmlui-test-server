package pathvars

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type ParamVars []ParamVar

func (vars ParamVars) Map() (m map[common.Identifier]ParamVar) {
	m = make(map[common.Identifier]ParamVar)
	for _, v := range vars {
		m[v.Name] = v
	}
	return m
}

type ParamVar struct {
	NameSpecProps
	Type        PVDataType
	Constraints []Constraint
	UseType     ParamUseType
}

var paramVarRegex = regexp.MustCompile(`\{.+?}`)

func ParseParamsInURLPath(path common.URLPath) (vars []ParamVar, err error) {
	var errs []error
	var dt PVDataType
	var props *NameSpecProps
	ut := PathUseType
	qPos := strings.Index(string(path), "?")
	matches := paramVarRegex.FindAllStringSubmatchIndex(string(path), -1)
	for i := 0; i < len(matches); i++ {
		var cs []Constraint
		varSpec := string(path[matches[i][0]+1 : matches[i][1]-1])
		parts := strings.SplitN(varSpec, ":", 3)
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		props, err = ParseNameSpecProps(parts[0])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		switch {
		case len(parts) == 1 || parts[1] == "":
			dt = InferDataTypeFromName(parts[0])
		case len(parts) > 1:
			dt, err = ParsePVDataType(parts[1])
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		if dt == UnspecifiedDataType {
			errs = append(errs, ErrInvalidNameSpec, ErrInvalidParameterType,
				fmt.Errorf("parameter_var=%s", varSpec))
			continue
		}
		if matches[i][0] > qPos {
			ut = PathUseType
		}
		if len(parts) > 2 {
			cs, err = ParseConstraints(parts[2], dt)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
		vars = append(vars, ParamVar{
			NameSpecProps: *props,
			Type:          dt,
			UseType:       ut,
			Constraints:   cs,
		})
	}
	return vars, errors.Join(errs...)
}
