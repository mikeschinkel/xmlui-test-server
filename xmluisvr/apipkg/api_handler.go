package apipkg

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

// ErrFailedToReadQueryFile indicates that an SQL file referenced by an endpoint could not be read.
var ErrFailedToReadQueryFile = errors.New("failed to read query file")

// loadAPIQuery loads the SQL query for an endpoint, either from inline configuration
// or from an external SQL file. If a query file is specified, it reads the file
// relative to the API configuration file's directory.
func (api *API) loadAPIQuery(ep *Endpoint) (q common.QueryString, err error) {
	var qf common.Filepath

	dir := common.DirPath(filepath.Dir(string(api.SourceFile)))
	q, qf, err = ep.GetQuery(dir)

	// Check if SQL should be loaded from a file
	if qf != "" {
		api.Info("Loaded query from file: %s", qf)
	}
	return q, err
}

// HandleAPIFunc returns an HTTP handler function that processes API requests.
// The handler:
//  1. Matches the request path against configured endpoints
//  2. Extracts path parameters and request body
//  3. Loads the SQL query for the matched endpoint
//  4. Executes the query against the database
//  5. Returns the results as JSON
//
// Returns 404 for unmatched routes and 500 for server errors.
func (api *API) HandleAPIFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var result pathvars.MatchResult
		var err error

		cliutil.Printf("APIConfig: %s %s", r.Method, r.URL.Path)

		if api == nil {
			// IS THIS EVEN NEEDED?
			common.SendErrorResponse(w, "APIConfig route not found; no APIConfig was loaded", http.StatusNotFound)
			return
		}

		// Find the matching endpoint
		result, err = api.Router.Match(r)
		if errors.Is(err, pathvars.ErrNoMatch) {
			api.Errorf("%s %s not matched: %v\n", r.Method, r.URL.Path, err.Error())
			http.NotFound(w, r)
			return
		}
		if err != nil {
			api.internalServerError(result, w, r, err, "Unexpected error while attempting to match")
			return
		}
		if result.Index < 0 || len(api.Endpoints) <= result.Index {
			api.internalServerError(result, w, r, err, "Unexpected match index out of bounds while attempting to match")
			return
		}
		endpoint := api.Endpoints[result.Index]

		// Extract body JSON if present
		bodyJSON, err := extractBodyJSON(r)
		if err != nil {
			log.Printf("Warning: Failed to parse request body as JSON: %v", err)
		}
		common.Noop(bodyJSON)

		// Prepare SQL query
		query, err := api.loadAPIQuery(endpoint)
		if err != nil {
			api.internalServerError(result, w, r, err, "Failed to load API query")
			return
		}

		// Replace named parameters with ? placeholders and build params array
		//queryParams := extractQueryParams(endpoint, qpArgs{
		//	params:   result.ParamsMap(),
		//	bodyJSON: bodyJSON,
		//})
		// Execute the query
		queryParams := []any{}
		dbResult, err := dbpkg.ExecuteQuery(ctx, db, query, queryParams)
		if err != nil {
			common.SendErrorResponse(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		// Return response
		common.SendJSONResponse(w, r, dbResult, http.StatusOK)
	}
}

// internalServerError logs an error and sends a 500 response to the client.
// It logs both to the CLI output and structured logger with context information.
func (api *API) internalServerError(mr pathvars.MatchResult, w http.ResponseWriter, r *http.Request, err error, msg string) {
	api.Errorf("%s %s %s: %v",
		r.Method,
		r.URL.Path,
		err,
	)
	api.Error(msg,
		"method", r.Method,
		"url_path", r.URL.Path,
		"match_result", mr,
		"error", err,
	)
	common.SendErrorResponse(w, msg, http.StatusInternalServerError)
}
