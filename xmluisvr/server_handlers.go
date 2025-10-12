package xmluisvr

import (
	"bytes"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiresp"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
)

func (svr *Server) handleRootFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svr.Printf("Request: %s\n", r.URL.Path)
		if r.URL.Path != "/" {
			svr.serveFile(w, r, common.Filepath("."+r.URL.Path))
			return
		}
		svr.serveFile(w, r, "./index.html")
	}
}

func (svr *Server) serveFile(w http.ResponseWriter, r *http.Request, filePath common.Filepath) {
	svr.Writer.Printf("Trying to serve: %s\n", filePath)
	err := common.CheckFileExists(filePath)
	switch {
	case errors.Is(os.ErrNotExist, err):
		svr.Writer.Errorf("File not found\n")
		http.NotFound(w, r)
	case errors.Is(ErrPathIsDir, err):
		svr.serveFile(w, r, common.Filepath(fmt.Sprintf("%s/index.html", filePath)))
	default:
		// TODO Make this safe from path traversal exploit
		http.ServeFile(w, r, filepath.Join(string(svr.api.Webroot), string(filePath)))
	}
}

// Handle direct SQL query requests
func (svr *Server) handleQueryFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		svr.V2().Printf("Query request: %svr\n", r.URL.Path)

		args := apipkg.HandlerHelperArgs{
			HTTPRequest: r,
			Database:    db,
			APIResponse: apiresp.NewResponse(apiresp.ResponseArgs{
				HTTPWriter: w,
				Request:    r,
				CLIWriter:  svr.Writer,
				Logger:     svr.Logger,
			}),
		}

		err = svr.checkUntrustedQueriesAuthorization(args)
		if err != nil {
			goto end
		}

		args.RequestBody, err = svr.getHTTPBody(args)
		if err != nil {
			goto end
		}

		args.DBQuery, args.QueryValues, err = svr.getDBQuery(args)
		if err != nil {
			goto end
		}

		// Now call the SQL query
		args.QueryResult, err = svr.api.GetQueryResult(ctx, args)
		if err != nil {
			goto end
		}

		// Now get the response content
		args.Content, err = svr.api.GetResponseContent(args)
		if err != nil {
			goto end
		}

		// Finally, send the success response. `args.Content` is expected to
		// contain the value to return as JSON.
		svr.api.SendSuccessResponse(args.SendResponseArgs(nil))
	end:
		if err != nil {
			// Or, send the error if it was an error This assumes that err will have been
			// joined with a `ResponsePayload` — which by declaration is also an `error` —
			// one example being rfc9457.Response.
			svr.api.SendErrorResponse(args.SendResponseArgs(err))
		}
		return
	}
}

// Handle proxy requests
func (svr *Server) checkUntrustedQueriesAuthorization(args apipkg.HandlerHelperArgs) (err error) {
	if !svr.options.AllowUntrustedQueries {
		// Should PresentationStyle not return 404 instead, or it 501 still valid? 501 is
		// probably valid since this is not a production server and leaking info is not a
		// big concern for local development and testing.
		err = errors.Join(
			ErrUnauthorizedEndpointAccess,
			apiresp.UnauthorizedPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Detail: fmt.Sprintf(
					apiresp.AllowUntrustedQueriesErrorDetail,
					args.APIEndpointRequested(),
				),
				Suggestion: fmt.Sprintf(
					apiresp.AllowUntrustedQueriesErrorSuggestion,
					cfgldr.AllowUntrustedQueriesFlag,
					args.APIEndpointRequested(),
				),
			}),
		)
	}
	return err
}

var (
	ErrFailedToUnmarshalJSON                 = errors.New("failed to unmarshal JSON")
	ErrInvalidDBQueryString                  = errors.New("invalid database query string; failed to parse")
	ErrFailedToGetDBQueryFromHTTPRequestBody = errors.New("failed to get database query from HTTP request body")
	ErrInvalidURL                            = errors.New("invalid URL")
	ErrMissingHostAfterProxySegment          = errors.New("missing host after proxy segment")
	ErrInvalidProxyTargetHost                = errors.New("invalid target proxy host")
)

func (svr *Server) getHTTPBody(args apipkg.HandlerHelperArgs) (body bytes.Buffer, err error) {
	errorStyle := svr.options.ErrorStyle

	// Use io.TeeReader to log the body while still allowing it to be read
	teeReader := io.TeeReader(args.HTTPRequest.Body, &body)
	// Read the body into a buffer
	_, err = io.ReadAll(teeReader)
	if err != nil {
		err = errors.Join(
			apiresp.ErrFailedToReadHTTPRequestBody,
			apiresp.InternalServerErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Location: apiresp.BodyLocation,
				Suggestion: errorStyle.ErrorMessage(
					fmt.Sprintf(apiresp.TryRestartingTheServerOrFileOnGithub, apiresp.ReportOnGithubMessageFunc()),
					apiresp.ErrFailedToReadHTTPRequestBody.Error(),
					err,
				),
			}),
			err,
		)
		goto end
	}
end:
	return body, err
}

func (svr *Server) getDBQuery(args apipkg.HandlerHelperArgs) (qs dbqvars.QueryString, values []any, err error) {
	// Decode the body into the queryRequest struct
	var req struct {
		Query  string `json:"db_query"`
		Values []any  `json:"parameters"`
	}
	err = jsonv2.UnmarshalRead(&args.RequestBody, &req)
	if err != nil {
		err = errors.Join(
			ErrFailedToGetDBQueryFromHTTPRequestBody,
			ErrFailedToUnmarshalJSON,
			apiresp.InvalidBodyFormatErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Detail: fmt.Sprintf("%s; %s",
					ErrFailedToGetDBQueryFromHTTPRequestBody.Error(),
					ErrFailedToUnmarshalJSON.Error(),
				),
				Suggestion: apiresp.EnsureYourHTTPRequestBodyContainsAValidDBQueryJSON,
			}),
		)
		goto end
	}
	values = req.Values
	qs, err = args.Database.ParseQueryString(req.Query)
	if err != nil {
		err = errors.Join(
			ErrFailedToGetDBQueryFromHTTPRequestBody,
			ErrInvalidDBQueryString,
			apiresp.InvalidBodyFormatErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Detail:     svr.options.ErrorStyle.ErrorMessage("Invalid Database Query", fmt.Sprintf("Query=%s", req.Query), err),
				Suggestion: fmt.Sprintf(apiresp.EnsureYourDBQueryIsValidForDB, svr.displayDBTypeName()),
			}),
		)
		goto end
	}
end:
	return qs, values, err
}

// Handle proxy requests
func (svr *Server) handleProxyFunc(method common.HTTPMethod) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		args := apipkg.HandlerHelperArgs{
			HTTPRequest: r,
			Database:    svr.db,
			APIResponse: apiresp.NewResponse(apiresp.ResponseArgs{
				HTTPWriter: w,
				Request:    r,
				CLIWriter:  svr.Writer,
				Logger:     svr.Logger,
			}),
		}

		args.TargetURL, args.URLPath, err = svr.getTargetURLAndPath(args)
		if err != nil {
			goto end
		}

		svr.createProxy(args).ServeHTTP(w, r) // do not mutate r beforehand

	end:
		if err != nil {
			// Or, send the error if it was an error This assumes that err will have been
			// joined with a `ResponsePayload` — which by declaration is also an `error` —
			// one example being rfc9457.Response.
			svr.api.SendErrorResponse(args.SendResponseArgs(err))
		}
		return
	}
}

func (svr *Server) getTargetURLAndPath(args apipkg.HandlerHelperArgs) (target *url.URL, up common.URLPath, err error) {
	var targetHost string

	path := args.HTTPRequest.URL.Path
	// ParseBytes "/proxy/<host>/<subpath...>?<query>"
	targetPath := strings.TrimPrefix(path, "/proxy/")

	hostPart, path, _ := strings.Cut(targetPath, "/")
	if hostPart == "" {
		err = errors.Join(
			ErrInvalidURL,
			ErrMissingHostAfterProxySegment,
			apiresp.InvalidURLFormatErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Location:   apiresp.PathLocation,
				Detail:     fmt.Sprintf(`Invalid URL format; got %s`, path),
				Suggestion: apiresp.EnsureURLBeginsWithPrefix,
			}),
		)
		goto end
	}
	up = common.URLPath(path)

	target, err = url.Parse(fmt.Sprintf("https://%s", hostPart))
	targetHost = "'No host provided'"
	if target != nil && target.Host != "" {
		targetHost = target.Host
	}
	if err != nil {
		err = errors.Join(
			ErrInvalidURL,
			ErrInvalidProxyTargetHost,
			apiresp.InvalidURLFormatErrorPayload(args.HTTPRequest, apiresp.PayloadArgs{
				Location:   apiresp.PathLocation,
				Detail:     fmt.Sprintf(`Invalid URL format for proxy target host; got %s`, targetHost),
				Suggestion: apiresp.EnsureURLBeginsWithPrefix,
			}),
		)
		goto end
	}

end:
	return target, up, err
}

func (svr *Server) createProxy(args apipkg.HandlerHelperArgs) (proxy *httputil.ReverseProxy) {
	proxy = httputil.NewSingleHostReverseProxy(args.TargetURL)

	// Build a Director that *only* mutates the outbound request.
	proxy.Director = svr.proxyDirectorFunc(proxy.Director, proxy, args.HTTPRequest, args)
	// Give yourself visibility vs “mystery crash”
	proxy.ErrorHandler = svr.proxyErrorHandlerFunc(args.TargetURL)

	// (Optional) Hardened Transport (timeouts, no HTTP/2 if you suspect issues, etc.)
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// DisableCompression:  true, // sometimes useful when chasing bugs
	}

	svr.Writer.Printf("Proxying: %s %s\n", args.HTTPRequest.Method, args.TargetURL.String())

	return proxy
}

func (svr *Server) proxyErrorHandlerFunc(targetURL *url.URL) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		// Log and convert to a 502 (or 504 on timeout)
		status := http.StatusBadGateway
		var nErr net.Error
		if errors.As(err, &nErr) && nErr.Timeout() {
			status = http.StatusGatewayTimeout
		}
		svr.Errorf("Proxy error for https://%s%s; %v\n", targetURL.Host, req.URL.Path, err)
		svr.Error("Proxy error",
			"target_host", targetURL.Host,
			"url_path", req.URL.Path,
			"error", err,
		)
		http.Error(w, http.StatusText(status), status)
	}
}

func (svr *Server) proxyDirectorFunc(priorDirector func(*http.Request), proxy *httputil.ReverseProxy, in *http.Request, args apipkg.HandlerHelperArgs) func(*http.Request) {
	return func(out *http.Request) {
		in := args.HTTPRequest
		path := args.URLPath
		targetURL := args.TargetURL

		// Start with stdlib’s defaults.
		priorDirector(out)

		if len(path) >= 1 && path[0] != '/' {
			path = common.URLPath("/" + string(path))
		}
		if path == "" {
			path = "/"
		}

		// Then apply our mapping.
		out.URL.Path = string(path)
		out.URL.RawQuery = in.URL.RawQuery
		out.URL.Scheme = targetURL.Scheme
		out.URL.Host = targetURL.Host

		// Host header to upstream (avoid surprises)
		out.Host = targetURL.Host

		svr.Printf("Proxying %s to %s\n", in.URL.Path, targetURL.Host)

		// Forward the client IP chain
		out.Header.Set("X-Forwarded-Host", in.Host)
		out.Header.Set("X-Forwarded-Proto", InboundProxyProtocol)
		xff := out.Header.Get("X-Forwarded-For")
		ip, _, err := net.SplitHostPort(in.RemoteAddr)
		if err != nil {
			goto end
		}
		if xff != "" {
			ip = xff + ", " + ip
		}
		out.Header.Set("X-Forwarded-For", ip)
	end:
	}
}
