package dbpkg

import (
	"strings"

	"github.com/xmlui-org/localdev/xmluisvr/common"
)

type MultipartQuery struct {
	QuerySources []*QuerySource
}

func NewMultipartQuery() *MultipartQuery {
	return &MultipartQuery{
		QuerySources: make([]*QuerySource, 0),
	}
}

func (mpq *MultipartQuery) AddQuerySource(qs *QuerySource) {
	mpq.QuerySources = append(mpq.QuerySources, qs)
}

func (mpq *MultipartQuery) HasQueries() (has bool) {
	if len(mpq.QuerySources) == 0 {
		goto end
	}
	for _, qs := range mpq.QuerySources {
		if qs.Source == "" {
			continue
		}
		has = true
		goto end
	}
end:
	return has
}

func (mpq *MultipartQuery) Source() (src common.QueryString) {
	sb := strings.Builder{}
	for _, qs := range mpq.QuerySources {
		sb.WriteString(string(qs.Source))
		sb.WriteByte('\n')
	}
	return common.QueryString(sb.String())
}
