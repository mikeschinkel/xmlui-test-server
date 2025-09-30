package dbpkg

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

var mutex sync.Mutex

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

type QuerySource struct {
	Lines    [2]int
	Source   common.QueryString
	Filepath common.Filepath
}

func NewQuerySource(start, end int, src common.QueryString, fp common.Filepath) *QuerySource {
	return &QuerySource{
		Lines:    [2]int{start, end},
		Source:   src,
		Filepath: fp,
	}
}

type QueryResult []map[string]any

func formatParams(db Database, params []any) (out []string) {
	fn := db.GetFormatParamFunc()
	out = make([]string, len(params))
	for i, p := range params {
		out[i] = fmt.Sprintf("%s=%v", fn(i), p)
	}
	return out
}

// ExecuteQuery and return results as a map of any
func ExecuteQuery(ctx Context, db Database, query common.QueryString, params []any) (result QueryResult, err error) {
	var rows *sql.Rows
	var columns []string

	result = make(QueryResult, 0)

	mutex.Lock()
	defer mutex.Unlock()

	// Log the SQL query (just once)
	cliutil.Printf("Query: %s %v\n", query, formatParams(db, params))

	// Execute the query
	rows, err = db.Query(ctx, string(query), params...)
	if err != nil {
		goto end
	}
	defer common.CloseOrLog(rows)

	// Get column information
	columns, err = rows.Columns()
	if err != nil {
		goto end
	}

	// Process result rows
	for rows.Next() {
		// Create values slice with appropriate length
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into values
		err := rows.Scan(valuePtrs...)
		if err != nil {
			goto end
		}

		// Create a map for this row
		entry := make(map[string]any)
		for i, col := range columns {
			var v any
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}

		// Add the row to the result
		result = append(result, entry)
	}

	// Check for errors after iteration
	err = rows.Err()

end:
	return result, err
}
