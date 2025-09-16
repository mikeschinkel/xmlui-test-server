package apipkg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

func (api *API) HandleAPIFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var result []map[string]any

		cliutil.Errorf("APIConfig: %s %s", r.Method, r.URL.Path)

		if api == nil {
			// IS THIS EVEN NEEDED?
			common.SendErrorResponse(w, "APIConfig route not found; no APIConfig was loaded", http.StatusNotFound)
			return
		}

		// Find the matching endpoint
		endpoint, pathParams := api.findMatchingEndpoint(common.URLPath(r.URL.Path))
		if endpoint == nil {
			http.NotFound(w, r)
			return
		}

		// Extract parameters
		urlParams := extractURLParams(r)

		bodyJSON, err := extractBodyJSON(r)
		if err != nil {
			log.Printf("Warning: Failed to parse request body as JSON: %v", err)
		}

		// Prepare SQL query
		query := ""

		// Check if SQL should be loaded from a file
		if endpoint.QueryFile == "" {
			// Use the inline SQL from the APIConfig definition
			query = endpoint.Query
		} else {
			// Determine the APIConfig description file's directory to make relative paths work

			// Build the SQL file path relative to the APIConfig description file
			queryFile := filepath.Join(
				filepath.Dir(string(api.SourceFile)),
				string(endpoint.QueryFile),
			)
			log.Printf("Loading SQL from file: %s", queryFile)

			// Read the SQL file
			queryBytes, err := os.ReadFile(queryFile)
			if err != nil {
				common.SendErrorResponse(w, fmt.Sprintf("Failed to read SQL file: %v", err), http.StatusInternalServerError)
				return
			}

			// Use the file contents as the SQL query
			query = string(queryBytes)
		}

		// Replace named parameters with ? placeholders and build params array
		queryParams := extractQueryParams(endpoint, qpArgs{
			pathParams: pathParams,
			urlParams:  urlParams,
			bodyJSON:   bodyJSON,
		})

		// Execute the query
		result, err = dbpkg.ExecuteQuery(ctx, db, query, queryParams)
		if err != nil {
			common.SendErrorResponse(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		// Return response
		common.SendJSONResponse(w, r, result, http.StatusOK)
	}
}
