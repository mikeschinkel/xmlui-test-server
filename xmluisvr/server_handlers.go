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

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

func (s *Server) handleRootFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Printf("Request: %s\n", r.URL.Path)
		if r.URL.Path != "/" {
			s.serveFile(w, r, common.Filepath("."+r.URL.Path))
			return
		}
		s.serveFile(w, r, "./index.html")
	}
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, filePath common.Filepath) {
	s.Writer.Printf("Trying to serve: %s\n", filePath)
	err := common.CheckFileExists(filePath)
	switch {
	case errors.Is(os.ErrNotExist, err):
		s.Writer.Errorf("File not found\n")
		http.NotFound(w, r)
	case errors.Is(ErrPathIsDir, err):
		s.serveFile(w, r, common.Filepath(fmt.Sprintf("%s/index.html", filePath)))
	default:
		// TODO Make this safe from path traversal exploit
		http.ServeFile(w, r, filepath.Join(string(s.api.Webroot), string(filePath)))
	}
}

// Handle direct SQL query requests
func (s *Server) handleQueryFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.V2().Printf("Query request: %s", r.URL.Path)

		if !s.options.AllowUntrustedQueries {
			s.WarnError("Query disallowed!")
			common.SendErrorResponse(w, "Currently not allowing untrusted database queries to run", http.StatusNotImplemented)
			return
		}
		// Use io.TeeReader to log the body while still allowing it to be read
		var bodyBuffer bytes.Buffer
		teeReader := io.TeeReader(r.Body, &bodyBuffer)

		// Read the body into a buffer
		_, err := io.ReadAll(teeReader)
		if err != nil {
			common.SendErrorResponse(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		query := string(bodyBuffer.Bytes())
		s.V3().InfoPrint("Database query submitted.",
			"requestor_ip", r.RemoteAddr,
			"query", strings.Replace(query, "\n", " ", -1),
		)

		// Decode the body into the queryRequest struct
		var req struct {
			SQL    string `json:"sql"`
			Params []any  `json:"params"`
		}
		err = jsonv2.UnmarshalRead(&bodyBuffer, &req)
		if err != nil {
			common.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		qs, err := db.ParseQueryString(req.SQL)
		if err != nil {
			common.SendErrorResponse(w, "failed to parse database query", http.StatusInternalServerError)
			return
		}

		// Execute the query
		result, err := dbpkg.ExecuteQuery(ctx, db, qs, req.Params)
		if err != nil {
			common.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return response
		common.SendJSONResponse(w, r, result, http.StatusOK)
	}
}

// Handle proxy requests
func (s *Server) handleProxyFunc(method common.HTTPMethod) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse "/proxy/<host>/<subpath...>?<query>"
		targetPath := strings.TrimPrefix(r.URL.Path, "/proxy/")
		hostPart, rest, _ := strings.Cut(targetPath, "/")
		if hostPart == "" {
			s.Errorf("missing host after /proxy/")
			http.Error(w, "missing host after /proxy/", http.StatusBadRequest)
			return
		}

		// (Optional but wise) guard against Server-Side Request Forgery (SSRF) / illegal hosts
		// if !s.allowedHost(hostPart) { http.Error(...); return }

		// Construct a "bare" target with no path so the Director won't double up paths.
		rawTarget := "https://" + hostPart
		targetURL, err := url.Parse(rawTarget)
		if err != nil || targetURL.Host == "" {
			s.Errorf("invalid target host: %s\n", hostPart)
			http.Error(w, "invalid target host: "+hostPart, http.StatusBadRequest)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Build a Director that *only* mutates the outbound request.
		proxy.Director = s.proxyDirectorFunc(proxy.Director, proxy, r, targetURL, rest)
		// Give yourself visibility vs “mystery crash”
		proxy.ErrorHandler = s.proxyErrorHandlerFunc(targetURL)

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
		proxy.ServeHTTP(w, r) // do not mutate r beforehand

	}
}

func (s *Server) proxyErrorHandlerFunc(targetURL *url.URL) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		// Log and convert to a 502 (or 504 on timeout)
		status := http.StatusBadGateway
		var nErr net.Error
		if errors.As(err, &nErr) && nErr.Timeout() {
			status = http.StatusGatewayTimeout
		}
		s.Errorf("Proxy error for https://%s%s; %v\n", targetURL.Host, req.URL.Path, err)
		s.Error("Proxy error",
			"target_host", targetURL.Host,
			"url_path", req.URL.Path,
			"error", err,
		)
		http.Error(w, http.StatusText(status), status)
	}
}

func (s *Server) proxyDirectorFunc(priorDirector func(*http.Request), proxy *httputil.ReverseProxy, in *http.Request, targetURL *url.URL, rest string) func(*http.Request) {
	return func(out *http.Request) {
		// Start with stdlib’s defaults.
		priorDirector(out)

		// Then apply our mapping.
		subPath := "/"
		if rest != "" {
			subPath = "/" + rest
		}
		out.URL.Path = subPath
		out.URL.RawQuery = in.URL.RawQuery
		out.URL.Scheme = targetURL.Scheme
		out.URL.Host = targetURL.Host

		// Host header to upstream (avoid surprises)
		out.Host = targetURL.Host

		s.Printf("Proxying %s to %s\n", in.URL.Path, targetURL.Host)

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
