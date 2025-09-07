package apipkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

type API struct {
	Name        string
	Webroot     common.Filepath
	SourceFile  common.Filepath // Source filepath where the API file is defined
	BasePath    common.URLPath
	Endpoints   []*Endpoint
	Verbose     bool
	CLIWriter   cliutil.Writer
	Logger      *slog.Logger
	pathRegexps map[common.URLPath]*regexp.Regexp
	initialized bool
}

type APIArgs struct {
	Name       string
	Webroot    common.Filepath
	SourceFile common.Filepath // Source filepath where the API file is defined
	BasePath   common.URLPath
	Endpoints  []*Endpoint
	Verbose    bool
	CLIWriter  cliutil.Writer
	Logger     *slog.Logger
}

func NewAPI(args APIArgs) (api *API) {
	return &API{
		Name:        args.Name,
		Webroot:     args.Webroot,
		SourceFile:  args.SourceFile,
		BasePath:    args.BasePath,
		Endpoints:   args.Endpoints,
		Verbose:     args.Verbose,
		CLIWriter:   args.CLIWriter,
		Logger:      args.Logger,
		pathRegexps: make(map[common.URLPath]*regexp.Regexp),
	}
}

func MakeAPIArgs(cfg cfgldr.APIConfig, w cliutil.Writer, l *slog.Logger) (apiArgs APIArgs, err error) {
	var basePath common.URLPath
	var sourceFile common.Filepath
	var webroot common.Filepath
	var endpoints []*Endpoint

	cfgV2 := cfg.(*cfgldr.APIConfigV2)
	basePath, err = common.ParseURLPath(cfgV2.BasePath)
	if err != nil {
		goto end
	}
	webroot, err = common.ParseFilepath(cfgV2.Webroot)
	if err != nil {
		goto end
	}
	sourceFile, err = common.ParseFilepath(cfgV2.SourceFile)
	if err != nil {
		goto end
	}
	endpoints, err = ParseEndpoints(cfgV2.Endpoints)
	if err != nil {
		goto end
	}
	apiArgs = APIArgs{
		Name:       cfgV2.Name,
		Webroot:    webroot,
		SourceFile: sourceFile,
		BasePath:   basePath,
		Endpoints:  endpoints,
		CLIWriter:  w,
		Logger:     l,
	}
end:
	return apiArgs, err
}

//func (api *API) String() string {
//	return fmt.Sprintf("URL Path: %s (Source: %s)", api.Name, api.SourceFile)
//}

func (api *API) Initialize(_ context.Context) (err error) {
	if api.initialized {
		goto end
	}
	err = api.compileEndpoints()
	api.initialized = true
end:
	return err
}

// Find the matching endpoint for a request path
func (api *API) compileEndpoints() error {
	var errs []error
	for _, ep := range api.Endpoints {
		re, err := common.CompileURLPathToRegexp(ep.Endpoint)
		errs = append(errs, err)
		api.pathRegexps[ep.Endpoint] = re
	}
	return errors.Join(errs...)
}

func (api *API) HandleAPIFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var result []map[string]any

		cliutil.Errorf("API: %s %s", r.Method, r.URL.Path)

		if api == nil {
			// IS THIS EVEN NEEDED?
			common.SendErrorResponse(w, "API route not found; no API was loaded", http.StatusNotFound)
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
			// Use the inline SQL from the API definition
			query = endpoint.Query
		} else {
			// Determine the API description file's directory to make relative paths work

			// Build the SQL file path relative to the API description file
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

type qpArgs struct {
	pathParams map[common.Identifier]string
	urlParams  map[common.Identifier]string
	bodyJSON   map[common.Identifier]any
}

// Extract query parameters from database query
func extractQueryParams(endpoint *Endpoint, args qpArgs) (queryParams []any) {
	var query string
	for paramName := range endpoint.Params {
		colonName := string(":" + paramName)
		// Check path params first, then query params, then body params
		if value, ok := args.pathParams[paramName]; ok {
			queryParams = append(queryParams, value)
			query = strings.Replace(query, colonName, "?", 1)
			continue
		}

		if value, ok := args.urlParams[paramName]; ok {
			queryParams = append(queryParams, value)
			query = strings.Replace(query, colonName, "?", 1)
			continue
		}

		value, ok := args.bodyJSON[paramName]
		if ok {
			queryParams = append(queryParams, value)
			query = strings.Replace(query, colonName, "?", 1)
			continue
		}

		// Parameter not found, add nil
		queryParams = append(queryParams, nil)
		query = strings.Replace(query, colonName, "?", 1)

	}
	return queryParams
}

// Find the matching endpoint for a request path
func (api *API) findMatchingEndpoint(requestPath common.URLPath) (endpoint *Endpoint, params map[common.Identifier]string) {

	rp := string(requestPath)
	// Strip base path if present
	basePath := string(api.BasePath)
	if basePath != "" && strings.HasPrefix(rp, basePath) {
		rp = strings.TrimPrefix(rp, basePath)
		if rp == "" {
			rp = "/"
		}
	}

	// Normalize the path by removing trailing slashes
	normalizedPath := strings.TrimSuffix(rp, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// First try exact match with normalized path
	for _, ep := range api.Endpoints {
		re, exists := api.pathRegexps[ep.Endpoint]
		if !exists {
			// This shouldn't happen as we precompile all regexps
			cliutil.Printf("Warning: No regexp for path %s", ep.Endpoint)
			logger.Warn("No regexp for endpoint path", "endpoint_path", ep.Endpoint)
			continue
		}

		if !re.MatchString(normalizedPath) {
			continue
		}
		endpoint = ep
		params = extractPathParams(common.URLPath(normalizedPath), ep.Endpoint, re)
		goto end
	}

	// If we reach here, try matching with the original path as a fallback
	if normalizedPath != rp {
		for _, ep := range api.Endpoints {
			re, exists := api.pathRegexps[ep.Endpoint]
			if !exists {
				continue
			}

			if !re.MatchString(rp) {
				continue
			}

			endpoint = ep
			params = extractPathParams(common.URLPath(rp), ep.Endpoint, re)
			goto end
		}
	}
end:
	return endpoint, params
}

// Extract url parameters from request URL
func extractURLParams(r *http.Request) map[common.Identifier]string {
	urlParams := make(map[common.Identifier]string)
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}
		urlParams[common.Identifier(key)] = values[0]
	}
	return urlParams
}

// Extract JSON body parameters from request
func extractBodyJSON(r *http.Request) (params map[common.Identifier]any, err error) {
	var buffer bytes.Buffer
	var jsonBytes []byte

	params = make(map[common.Identifier]any)

	if r.Body == nil {
		goto end
	}

	jsonBytes, err = io.ReadAll(io.TeeReader(r.Body, &buffer))
	if err != nil {
		goto end
	}

	if len(jsonBytes) > 0 {
		err = json.Unmarshal(jsonBytes, &params)
		if err != nil {
			goto end
		}
	}

	// Reset r.Body for potential future use
	r.Body = io.NopCloser(&buffer)

end:
	return params, err
}

// Extract path parameters from a URL based on the endpoint path template
// Example: extractPathParams("/clients/123", "/clients/:id") -> {"id": "123"}
func extractPathParams(requestPath, endpointPath common.URLPath, re *regexp.Regexp) map[common.Identifier]string {
	params := make(map[common.Identifier]string)

	// Extract param names from the path template
	paramNames := make([]string, 0)
	pathParts := strings.Split(string(endpointPath), "/")
	for _, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			paramNames = append(paramNames, part[1:])
		}
	}

	// Extract values using regexp
	matches := re.FindStringSubmatch(string(requestPath))
	if len(matches) <= 1 {
		goto end
	}
	// First match is the whole string, subsequent matches are capture groups
	for i, name := range paramNames {
		if i+1 >= len(matches) {
			continue
		}
		params[common.Identifier(name)] = matches[i+1]
	}
end:
	return params
}
