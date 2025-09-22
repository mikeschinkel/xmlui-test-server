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
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/pathvars"
)

type API struct {
	Name       string
	Webroot    common.Filepath
	SourceFile common.Filepath // Source filepath where the API file is defined
	BasePath   common.URLPath
	Endpoints  []*Endpoint
	Verbose    bool
	cliutil.WriterLogger
	Router      *pathvars.Router
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

type CreateAPIArgs struct {
	Config cfgldr.APIConfig
	Writer cliutil.Writer
	Logger *slog.Logger
}

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

//	func (api *APIConfig) String() string {
//		return fmt.Sprintf("URL Path: %s (Source: %s)", api.Name, api.SourceFile)
//	}
func (api *API) Initialize(_ context.Context) (err error) {
	if api.initialized {
		goto end
	}
	err = api.initializeRouter()
	api.initialized = true
end:
	return err
}

// Find the matching endpoint for a request path
func (api *API) initializeRouter() (err error) {
	var errs []error
	for _, ep := range api.Endpoints {
		err = api.Router.AddRoute(pathvars.PathSpec(ep.path), ep.PathVarsParameters())
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

type qpArgs struct {
	params   map[common.Identifier]string
	bodyJSON map[common.Identifier]any
}

// Extract query parameters from database query
func extractQueryParams(endpoint *Endpoint, args qpArgs) (queryParams []any) {
	var query string
	for _, param := range endpoint.Params {
		colonName := fmt.Sprintf(":%s", param.Name)
		// Check path params first, then query params, then body params
		if value, ok := args.params[param.Name]; ok {
			queryParams = append(queryParams, value)
			query = strings.Replace(query, colonName, "?", 1)
			continue
		}

		value, ok := args.bodyJSON[param.Name]
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
