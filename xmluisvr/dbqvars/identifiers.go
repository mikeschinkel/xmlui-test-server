package dbqvars

import (
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type Parameter struct {
	Name  common.Selector
	Index int
}

func (p Parameter) IsIdentifier() bool {
	return !strings.Contains(string(p.Name), ".")
} //Absolute or Relative

type Parameters []Parameter

func NewParameters(names ...common.Selector) (ps Parameters) {
	ps = make(Parameters, len(names))
	for i, name := range names {
		ps[i] = NewParameter(name, i+1)
	}
	return ps
}
func NewParameter(name common.Selector, index int) Parameter {
	return Parameter{
		Name:  name,
		Index: index,
	}
}

// Identifiers extracts slice of common.Identifier from a Parameters value (a
// slice of []Parameter)
func (ps Parameters) Identifiers() (ids []common.Identifier) {
	ids = make([]common.Identifier, len(ps))
	for i, p := range ps {
		if !p.IsIdentifier() {
			continue
		}
		ids[i] = common.Identifier(p.Name)
	}
	return ids
}

func (ps Parameters) DottedSelectors() (selectors []common.Selector) {
	selectors = make([]common.Selector, 0, len(ps))
	for i, p := range ps {
		if p.IsIdentifier() {
			continue
		}
		selectors[i] = common.Selector(p.Name)
	}
	return selectors
}
