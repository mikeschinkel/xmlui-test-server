// Package apipkg provides the API endpoint management system for xmlui-test-server.
//
// This package handles the configuration-driven API endpoint system that allows
// defining HTTP endpoints through JSON configuration files. It supports:
//
//   - Dynamic endpoint routing based on URL patterns
//   - Path parameter extraction and validation
//   - SQL query execution with parameter binding
//   - JSON request/response handling
//   - File-based or inline SQL queries
//   - Type checking and constraint validation
//
// # Configuration Format
//
// API endpoints are defined in JSON configuration files with this structure:
//
//	{
//	  "name": "My API",
//	  "basePath": "/api/v1",
//	  "webroot": "./public",
//	  "endpoints": [
//	    {
//	      "endpoint": "GET /users/:id",
//	      "query": "SELECT * FROM users WHERE id = :id",
//	      "params": ["id:integer"]
//	    }
//	  ]
//	}
//
// # Usage Example
//
//	cfg := &cfgldr.APIConfigV2{...}
//	api, err := apipkg.CreateAPI(apipkg.CreateAPIArgs{
//		Config: cfg,
//		Writer: writer,
//		Logger: logger,
//	})
//	if err != nil {
//		return err
//	}
//	err = api.Initialize(ctx)
//	handler := api.HandleAPIFunc(ctx, database)
package apipkg

import (
	"bytes"
	"context"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

// API represents a configured API instance with endpoints, routing, and metadata.
// It manages HTTP endpoints that execute SQL queries based on JSON configuration.
type API struct {
	Name                 string           // Human-readable name for the API
	Webroot              common.Filepath  // Root directory for static file serving
	SourceFile           common.Filepath  // Path to the configuration file
	BasePath             common.URLPath   // Common URL prefix for all endpoints
	Endpoints            []*Endpoint      // List of configured API endpoints
	Verbose              bool             // Enable verbose logging
	Router               *pathvars.Router // URL routing and path parameter extraction
	initialized          bool             // Whether Initialize() has been called
	cliutil.WriterLogger                  // Embedded logging functionality
}

// APIArgs contains the configuration needed to create a new API instance.
type APIArgs struct {
	Name       string          // API name
	Webroot    common.Filepath // Static file root directory
	SourceFile common.Filepath // Configuration file path
	BasePath   common.URLPath  // URL prefix for endpoints
	Endpoints  []*Endpoint     // Parsed endpoint configurations
	Verbose    bool            // Enable verbose output
	CLIWriter  cliutil.Writer  // CLI output writer
	Logger     *slog.Logger    // Structured logger
}

// CreateAPIArgs contains dependencies needed to create an API from configuration.
type CreateAPIArgs struct {
	Config cfgldr.APIConfig // Loaded API configuration
	Writer cliutil.Writer   // CLI writer for output
	Logger *slog.Logger     // Logger instance
}

// CreateAPI creates a new API instance from the provided configuration.
// It parses the configuration, validates settings, and creates endpoints.
// Currently only supports APIConfigV2 format.
func CreateAPI(args CreateAPIArgs) (api *API, err error) {
	var basePath common.URLPath
	var sourceFile common.Filepath
	var webroot common.Filepath
	var endpoints []*Endpoint

	cfg := args.Config

	cfgV2, ok := cfg.(*cfgldr.APIConfigV2)
	if !ok {
		panic(fmt.Sprintf("Cannot type assert APIConfig config value of type %T to type %T", cfg, (*cfgldr.APIConfigV2)(nil)))
	}
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
	api = NewAPI(APIArgs{
		Name:       cfgV2.Name,
		Webroot:    webroot,
		SourceFile: sourceFile,
		BasePath:   basePath,
		Endpoints:  endpoints,
		CLIWriter:  args.Writer,
		Logger:     args.Logger,
	})
end:
	return api, err
}

// NewAPI creates a new API instance with the provided arguments.
// The API is created with an empty router that must be initialized
// before use by calling Initialize().
func NewAPI(args APIArgs) (api *API) {
	return &API{
		Name:         args.Name,
		Webroot:      args.Webroot,
		SourceFile:   args.SourceFile,
		BasePath:     args.BasePath,
		Endpoints:    args.Endpoints,
		Verbose:      args.Verbose,
		Router:       pathvars.NewRouter(),
		WriterLogger: cliutil.NewWriterLogger(args.CLIWriter, args.Logger),
	}
}

// Initialize prepares the API for use by setting up the routing system.
// This method must be called before using HandleAPIFunc().
// It parses all endpoint path patterns and builds the internal router.
func (api *API) Initialize(_ context.Context) (err error) {
	if api.initialized {
		goto end
	}
	err = api.initializeRouter()
	api.initialized = true
end:
	return err
}

// initializeRouter configures the internal router with all endpoint patterns.
// It parses path variables from each endpoint and registers them with the router.
func (api *API) initializeRouter() (err error) {
	var errs []error
	for i, ep := range api.Endpoints {
		var pp []pathvars.Parameter
		pp, err = ep.ParsePathVarsParameters()
		if err != nil {
			errs = append(errs, err)
		}
		err = api.Router.AddRouteWithIndex(pathvars.PathSpec(ep.path), pp, i)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		err = errors.Join(errs...)
	}
	if err == nil {
		err = api.Router.Compile()
	}
	return err
}

// qpArgs contains parameters extracted from different sources for query building.
type qpArgs struct {
	params   map[pathvars.PVNameSpec]string // Path and query parameters
	bodyJSON map[common.Identifier]any      // JSON body parameters
}

// extractBodyJSON parses JSON from the request body into a parameter map.
// It uses a TeeReader to preserve the request body for potential future use.
// Returns an empty map if no JSON body is present or parsing fails.
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
		err = jsonv2.Unmarshal(jsonBytes, &params)
		if err != nil {
			goto end
		}
	}

	// Reset r.Body for potential future use
	r.Body = io.NopCloser(&buffer)

end:
	return params, err
}
