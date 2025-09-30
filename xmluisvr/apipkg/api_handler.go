package apipkg

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

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
		var dbq dbqvars.ParsedQuery
		var qs common.QueryString
		var endpoint *Endpoint
		var queryValues []any
		var dbResult dbpkg.QueryResult
		var notFound []common.Selector

		cliutil.Printf("API Request: %s %s\n", r.Method, r.URL.Path)

		if api == nil {
			// IS THIS EVEN NEEDED?
			common.SendErrorResponse(w, "APIConfig route not found; no APIConfig was loaded", http.StatusNotFound)
			goto end
		}

		// Find the matching endpoint
		result, err = api.Router.Match(r)
		if errors.Is(err, pathvars.ErrNoMatch) {
			api.Errorf("%s %s not matched: %v\n", r.Method, r.URL.Path, err.Error())
			http.NotFound(w, r)
			goto end
		}
		if err != nil {
			api.internalServerError(result, w, r, err, "Unexpected error while attempting to match")
			goto end
		}
		if result.Index < 0 || len(api.Endpoints) <= result.Index {
			api.internalServerError(result, w, r, err, "Unexpected match index out of bounds while attempting to match")
			goto end
		}
		endpoint = api.Endpoints[result.Index]

		queryValues, notFound, err = endpoint.GetParameterValues(r.Body)
		if err != nil {
			msg := "Database parameters not found"
			status := http.StatusBadRequest
			if len(notFound) == 0 {
				msg = "Unexpected database parameter error"
				status = http.StatusInternalServerError
			}
			_ = api.ErrorError(msg, "error", err)
			common.SendErrorResponse(w, fmt.Sprintf("%s: check the logs for details", msg), status)
			goto end
		}

		qs = endpoint.ParsedQuery.QueryString()
		dbResult, err = dbpkg.ExecuteQuery(ctx, db, qs, queryValues)
		if err != nil {
			// TODO: Response should not return err
			_ = api.ErrorError("Database error",
				"endpoint", endpoint.Endpoint(),
				"sql_query", dbq.QueryString(),
				"sql_params", dbq.Parameters(),
				"url_params", result.ParamsMap(),
				"error", err,
			)
			common.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
			goto end
		}
		if len(dbResult) == 0 && !result.Route.Cardinality.EmptyOk() {
			common.SendErrorResponse(w, "No results found", http.StatusNotFound)
			api.InfoPrint("No results found",
				"endpoint", endpoint.Endpoint(),
				"sql_query", dbq.QueryString(),
				"sql_params", dbq.Parameters(),
				"url_params", result.ParamsMap(),
			)
			goto end
		}
		// TODO Validate result types and column types

		// Return response
		common.SendJSONResponse(w, r, dbResult, http.StatusOK)
	end:
		return
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
