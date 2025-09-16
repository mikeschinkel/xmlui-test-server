package xmluisvr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/fsutil"
)

type API struct{}

type Server struct {
	db         dbpkg.Database
	api        *apipkg.API
	options    *common.Options
	port       common.ServerPort
	sourceFile common.Filepath
	mux        *http.ServeMux
}

type ServerArgs struct {
	Database   dbpkg.Database
	API        *apipkg.API
	Port       common.ServerPort
	SourceFile common.Filepath
	Options    *common.Options
}

func NewServer(args ServerArgs) *Server {
	if args.Port == 0 {
		args.Port = common.DefaultServerPort
	}
	return &Server{
		db:         args.Database,
		api:        args.API,
		port:       args.Port,
		options:    args.Options,
		sourceFile: args.SourceFile,
		mux:        http.NewServeMux(),
	}
}

func (s *Server) Initialize(ctx Context) (err error) {
	err = s.api.Initialize(ctx)
	if errors.Is(err, common.ErrNoAPIProvided) {
		cliutil.Printf("No APIConfig loaded")
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("failed to load APIConfig: %w", err)
		goto end
	}

	// Add URL routes
	s.addRoutes(ctx)

	err = s.db.Open(ctx)
	if err != nil {
		err = fmt.Errorf("failed to open database: %w", err)
		goto end
	}

end:
	return err
}

func (s *Server) ListenAndServe(_ Context) (err error) {
	cliutil.Printf("Listening on %s...\n", s.friendlyHost())
	return http.ListenAndServe(s.Host(), s.corsMiddleware(s.mux))
}

func (s *Server) showConfig() {

	cliutil.Printf("\n")
	cliutil.Printf("Configuration:\n")
	cliutil.Printf("- Current Dir:  %v\n", s.displayDir())
	cliutil.Printf("- Server:       %s\n", s.Host())
	cliutil.Printf("  - Webroot:    %v\n", s.displayWebroot())
	cliutil.Printf("  - Config:     %s\n", s.displayServerSourceFile())
	cliutil.Printf("- Database:     %s\n", s.databaseTypeName())
	cliutil.Printf("  - Connection: %v\n", s.databaseName())
	cliutil.Printf("  - Config:     %s\n", s.displayDBSourceFile())
	if s.api != nil {
		cliutil.Printf("- API:          %s\n", s.api.Name)
		cliutil.Printf("  - URL Path:   %s\n", s.api.BasePath)
		cliutil.Printf("  - Config:     %s\n", s.displayAPISourceFile())
	}
	if len(s.db.Extensions()) != 0 {
		cliutil.Printf("- Extension:   %s\n", s.extensionPaths())
	}
	if s.options.Verbose {
		cliutil.Printf("- Verbose:    true\n")
	}
	cliutil.Printf("\n")

}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Host() string {
	return fmt.Sprintf("%s:%d", common.LocalHostIP, s.port)
}
func (s *Server) Port() common.ServerPort {
	return s.port
}

func (s *Server) friendlyHost() string {
	return fmt.Sprintf("localhost:%d", s.port)
}

func (s *Server) addRoutes(ctx Context) {

	// Handle APIConfig routes first (to match /apiFile/* before static files)
	if s.api != nil {
		apiWebroot := string(s.api.Webroot)
		if !strings.HasSuffix(apiWebroot, "/") {
			apiWebroot += "/"
		}
		s.mux.HandleFunc("GET "+apiWebroot, s.api.HandleAPIFunc(ctx, s.db))
	}

	// Handle proxy next
	for _, method := range common.HTTPMethods {
		s.mux.HandleFunc(method+" /proxy/", s.handleProxy)
	}

	// Then handle query endpoint
	s.mux.HandleFunc("POST /query", s.handleQueryFunc(ctx, s.db))

	s.mux.HandleFunc("GET /", s.handleRoot())

}

func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliutil.Printf("Received request for: %s\n", r.URL.Path)
		if r.URL.Path != "/" {
			s.serveFile(w, r, common.Filepath("."+r.URL.Path))
			return
		}
		s.serveFile(w, r, "./index.html")
	}
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, filePath common.Filepath) {
	cliutil.Printf("Trying to serve: %s\n", filePath)
	err := common.CheckFileExists(filePath)
	switch {
	case errors.Is(os.ErrNotExist, err):
		cliutil.Errorf("File not found\n")
		http.NotFound(w, r)
	case errors.Is(ErrPathIsDir, err):
		s.serveFile(w, r, common.Filepath(fmt.Sprintf("%s/index.html", filePath)))
	default:
		// TODO Make this safe from path traversal exploit
		http.ServeFile(w, r, filepath.Join(string(s.api.Webroot), string(filePath)))
	}
}

func (s *Server) displayWebroot() (wr string) {
	// Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		wr = err.Error()
		goto end
	}
	if wd == "" {
		wr = "Working directory unavailable"
		goto end
	}
	wd = filepath.Join(wd, string(s.api.Webroot))
	wr = fsutil.HomeRelative(wd)
end:
	return wr
}

func (s *Server) displayAPISourceFile() string {
	return fsutil.HomeRelative(string(s.api.SourceFile))
}
func (s *Server) displayDBSourceFile() string {
	return fsutil.HomeRelative(string(s.db.SourceFile()))
}
func (s *Server) displayServerSourceFile() string {
	return fsutil.HomeRelative(string(s.sourceFile))
}

func (s *Server) displayDir() (d string) {
	// Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		d = err.Error()
		goto end
	}
	if wd == "" {
		d = "Working directory unavailable"
		goto end
	}
	d = fsutil.HomeRelative(wd)
end:
	return d
}

func (s *Server) databaseName() (name string) {
	if s.db == nil {
		return ""
	}
	return s.db.String()
}
func (s *Server) databaseTypeName() (name string) {
	return s.db.TypeName()
}

func (s *Server) extensionPaths() (paths string) {
	var sb strings.Builder
	for _, ext := range s.db.Extensions() {
		sb.WriteString(ext.Name())
		sb.WriteString(". ")
	}
	paths = sb.String()
	if len(paths) != 0 {
		paths = paths[:len(paths)-2]
	}
	return paths
}

// Handle APIConfig requests based on the APIConfig description

// Handle direct SQL query requests
func (s *Server) handleQueryFunc(ctx Context, db dbpkg.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Query: %s", r.URL.Path)

		// Use io.TeeReader to log the body while still allowing it to be read
		var bodyBuffer bytes.Buffer
		teeReader := io.TeeReader(r.Body, &bodyBuffer)

		// Read the body into a buffer
		_, err := io.ReadAll(teeReader)
		if err != nil {
			common.SendErrorResponse(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// Decode the body into the queryRequest struct
		var req struct {
			SQL    string `json:"sql"`
			Params []any  `json:"params"`
		}
		if !ask("Do we *REALLY* want to allow untrusted SQL to run?") {
			common.SendErrorResponse(w, "Currently not allowing untrusted SQL to run", http.StatusNotImplemented)
			return
		}
		err = json.NewDecoder(&bodyBuffer).Decode(&req)
		if err != nil {
			common.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Execute the query
		result, err := dbpkg.ExecuteQuery(ctx, db, req.SQL, req.Params)
		if err != nil {
			common.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return response
		common.SendJSONResponse(w, r, result, http.StatusOK)
	}
}
func ask(msg string) bool {
	cliutil.Errorf("%s\n", msg)
	return false
}

// Handle proxy requests
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Parse off the part after "/proxy/".
	targetPath := strings.TrimPrefix(r.URL.Path, "/proxy/")
	targetQuery := r.URL.RawQuery

	// 2. Split off the first segment as the actual host.
	pathParts := strings.SplitN(targetPath, "/", 2)
	hostPart := pathParts[0]

	// 3. The remainder is your path on that host.
	var subPath string
	if len(pathParts) > 1 {
		subPath = "/" + pathParts[1]
	} else {
		subPath = "/"
	}

	// 4. Construct a "bare" target with no path so the default Director won't double up paths.
	rawTarget := "https://" + hostPart
	targetURL, err := url.Parse(rawTarget)
	if err != nil {
		common.SendErrorResponse(w, "Invalid target URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 5. Create the reverse proxy.
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 6. Update the inbound request with subPath and query
	r.URL.Scheme = targetURL.Scheme
	r.URL.Host = targetURL.Host
	r.URL.Path = subPath
	r.URL.RawQuery = targetQuery

	// 7. (Optional) Reassign the host header to match target
	r.Host = targetURL.Host

	// 8. Finally, run the proxy
	proxy.ServeHTTP(w, r)
}
