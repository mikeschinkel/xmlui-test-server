package dbqvars

import (
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type Parameter string

func (p Parameter) IsIdentifier() bool {
	return !strings.Contains(string(p), ".")
} //Absolute or Relative

type Parameters []Parameter

// Identifiers extracts slice of common.Identifier from a Parameters value (a
// slice of []Parameter)
func (ps Parameters) Identifiers() (ids []common.Identifier) {
	ids = make([]common.Identifier, 0, len(ps))
	for i, id := range ps {
		if !id.IsIdentifier() {
			continue
		}
		ids[i] = common.Identifier(id)
	}
	return ids
}

func (ps Parameters) DottedSelectors() (selectors []common.Selector) {
	selectors = make([]common.Selector, 0, len(ps))
	for i, id := range ps {
		if id.IsIdentifier() {
			continue
		}
		selectors[i] = common.Selector(id)
	}
	return selectors
}
