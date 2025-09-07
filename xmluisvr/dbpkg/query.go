package dbpkg

import (
	"database/sql"
	"sync"

	"github.com/xmlui-org/xmluisvr/cliutil"
)

var mutex sync.Mutex

type QueryResult []map[string]any

// ExecuteQuery and return results as a map of any
func ExecuteQuery(ctx Context, db Database, query string, params []any) (result QueryResult, err error) {
	var rows *sql.Rows
	var columns []string

	result = make(QueryResult, 0)

	mutex.Lock()
	defer mutex.Unlock()

	// Log the SQL query (just once)
	cliutil.Printf("SQL: %s", query)

	// Execute the query
	rows, err = db.Query(ctx, query, params...)
	if err != nil {
		goto end
	}
	defer closeOrLog(rows)

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
