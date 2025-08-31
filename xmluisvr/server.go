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
	"regexp"
	"strings"
	"sync"

	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/dbutil"
	"github.com/xmlui-org/xmluisvr/fsutil"
)

type Server struct {
	db             dbutil.Database
	port           int
	apiDesc        *APIDescription
	connStr        string
	apiDescPath    string                    // Path to the API description file
	pathRegexps    map[string]*regexp.Regexp // Cache for compiled path regexps
	verbose        bool                      // Flag to enable/disable response logging
	dbType         dbutil.DatabaseType       // Type of database: "sqlite" or "postgres"
	extensionPaths []string
	mutex          sync.Mutex // Mutex to serialize DB access
	mux            *http.ServeMux
}

type ServerArgs struct {
	DatabaseType     dbutil.DatabaseType
	ConnectionString string
	ExtensionPaths   []string
	APIPath          string
	Verbose          bool
	Port             int
}

func NewServer(args ServerArgs) *Server {
	if args.Port == 0 {
		args.Port = DefaultPort
	}
	return &Server{
		port:           args.Port,
		apiDescPath:    args.APIPath,
		verbose:        args.Verbose,
		dbType:         args.DatabaseType,
		connStr:        args.ConnectionString,
		extensionPaths: args.ExtensionPaths,
		mutex:          sync.Mutex{},
		pathRegexps:    make(map[string]*regexp.Regexp),
		mux:            http.NewServeMux(),
	}
}

func (s *Server) Initialize() (err error) {

	// Load the API description if provided
	err = s.loadAPI()
	if err != nil {
		cliutil.Errorf("Failed to load API description file: %v", err)
		err = nil
	}

	// Add URL routes
	s.addRoutes()

	// Creates an instance of a Database object
	s.db, err = dbutil.CreateDatabase(s.dbType, dbutil.DatabaseArgs{
		DatabaseType:     s.dbType,
		ConnectionString: s.connStr,
		ExtensionPaths:   s.extensionPaths,
		CLIWriter:        cliutil.GetWriter(),
		Logger:           logger,
	})
	// Opens the database
	err = s.db.Open()

	return err
}

func (s *Server) ListenAndServe() (err error) {
	cliutil.Printf("Listening on %s...", s.friendlyHost())
	return http.ListenAndServe(s.host(), s.corsMiddleware(s.mux))
}

func (s *Server) showConfig() {

	cliutil.Printf("\n")
	cliutil.Printf("Configuration:\n")
	cliutil.Printf("- Host:      127.0.0.1\n")
	cliutil.Printf("- Port:      %d\n", s.port)
	cliutil.Printf("- DB Type:   %v\n", s.databaseTypeName())
	cliutil.Printf("- Database:  %v\n", s.databaseName())
	if s.apiDescPath != "" {
		cliutil.Printf("- API:       %s\n", s.apiPath())
	}
	if len(s.extensionPaths) != 0 {
		cliutil.Printf("- Extension: %s\n", s.extensionPath())
	}
	if s.verbose {
		cliutil.Printf("- Verbose:   true\n")
	}
	cliutil.Printf("- Directory: %v\n", s.displayDir())
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

func (s *Server) host() string {
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

func (s *Server) friendlyHost() string {
	return fmt.Sprintf("localhost:%d", s.port)
}

func (s *Server) addRoutes() {

	// Handle API routes first (to match /api/* before static files)
	if s.apiDesc != nil {
		apiBasePath := s.apiDesc.BasePath
		if !strings.HasSuffix(apiBasePath, "/") {
			apiBasePath += "/"
		}
		s.mux.HandleFunc(apiBasePath, s.handleAPI)
	}

	// Handle proxy next
	s.mux.HandleFunc("/proxy/", s.handleProxy)

	// Then handle query endpoint
	s.mux.HandleFunc("/query", s.handleQuery)

	s.mux.HandleFunc("/", s.handleRoot())
}

func (s *Server) handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliutil.Printf("Received request for: %s\n", r.URL.Path)
		if r.URL.Path != "/" {
			s.serveFile(w, r, "."+r.URL.Path)
			return
		}
		s.serveFile(w, r, "./index.html")
	}
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	cliutil.Printf("Trying to serve: %s\n", filePath)
	err := checkFileExists(filePath)
	switch {
	case errors.Is(os.ErrNotExist, err):
		cliutil.Errorf("File not found\n")
		http.NotFound(w, r)
	case errors.Is(ErrPathIsDir, err):
		s.serveFile(w, r, fmt.Sprintf("%s/index.html", filePath))
	default:
		http.ServeFile(w, r, filePath)
	}
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

func (s *Server) apiPath() string {
	return fsutil.HomeRelative(s.apiDescPath)
}

func (s *Server) databaseName() (name string) {
	if s.db == nil {
		return ""
	}
	return s.db.Name()
}
func (s *Server) databaseTypeName() (name string) {
	return s.dbType.String()
}

func (s *Server) extensionPath() string {
	return strings.Join(s.extensionPaths, ",")
}

func (s *Server) loadAPI() (err error) {
	// Load the API description if provided
	if s.apiDescPath == "" {
		goto end
	}
	err = checkFileExists(s.apiDescPath)
	if err != nil {
		cliutil.Errorf("Invalid API description file: %v", err)
		err = nil
		goto end
	}
	s.apiDesc, err = s.loadAPIDescription()
	if err != nil {
		cliutil.Errorf("Failed to load API description file: %v", err)
		err = nil
		goto end
	}

	cliutil.Printf("API description loaded successfully: %s (v%s)",
		s.apiDesc.Name,
		s.apiDesc.APIVersion,
	)

	// Precompile the path regexps for faster matching
	err = s.compileEndpoints()

end:
	return err
}

// Find the matching endpoint for a request path
func (s *Server) compileEndpoints() error {
	var errs []error
	for _, ep := range s.apiDesc.Endpoints {
		re, err := regexp.Compile(pathToRegexp(ep.Path))
		errs = append(errs, err)
		s.pathRegexps[ep.Path] = re
	}
	return errors.Join(errs...)
}

// Find the matching endpoint for a request path
func (s *Server) findMatchingEndpoint(requestPath string) (*EndpointDefinition, map[string]string) {
	if s.apiDesc == nil {
		return nil, nil
	}

	// Strip base path if present
	basePath := s.apiDesc.BasePath
	if basePath != "" && strings.HasPrefix(requestPath, basePath) {
		requestPath = strings.TrimPrefix(requestPath, basePath)
		if requestPath == "" {
			requestPath = "/"
		}
	}

	// Normalize the path by removing trailing slashes
	normalizedPath := strings.TrimSuffix(requestPath, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// First try exact match with normalized path
	for _, endpoint := range s.apiDesc.Endpoints {
		re, exists := s.pathRegexps[endpoint.Path]
		if !exists {
			// This shouldn't happen as we precompile all regexps
			log.Printf("Warning: No regexp for path %s", endpoint.Path)
			continue
		}

		if re.MatchString(normalizedPath) {
			params := extractPathParams(normalizedPath, endpoint.Path, re)
			return &endpoint, params
		}
	}

	// If we reach here, try matching with the original path as a fallback
	if normalizedPath != requestPath {
		for _, endpoint := range s.apiDesc.Endpoints {
			re, exists := s.pathRegexps[endpoint.Path]
			if !exists {
				continue
			}

			if re.MatchString(requestPath) {
				params := extractPathParams(requestPath, endpoint.Path, re)
				return &endpoint, params
			}
		}
	}

	return nil, nil
}

// Execute SQL query and return results as maps
func (s *Server) executeQuery(sqlQuery string, params []any) ([]map[string]any, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Log the SQL query (just once)
	log.Printf("SQL: %s", sqlQuery)

	// Handle PostgreSQL parameter placeholders ($1, $2, etc.) vs SQLite (?, ?, etc.)
	if s.dbType == "postgres" {
		// Replace ? with $1, $2, etc. for PostgreSQL
		for i := 1; i <= len(params); i++ {
			sqlQuery = strings.Replace(sqlQuery, "?", fmt.Sprintf("$%d", i), 1)
		}
	}

	// Execute the query
	rows, err := s.db.Query(sqlQuery, params...)
	if err != nil {
		return nil, err
	}
	defer closeOrLog(rows)

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Process result rows
	var result []map[string]any
	for rows.Next() {
		// Create values slice with appropriate length
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into values
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Log the response if enabled
	if s.verbose {
		// Response logging is done in sendJSONResponse
	}

	return result, nil
}

// Send JSON response with the given status code
func (s *Server) sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Generate JSON response
	responseJSON, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log the response if enabled - this is the ONLY place where responses should be logged
	if s.verbose {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, responseJSON, "", "  "); err != nil {
			log.Printf("Error prettifying JSON for logging: %v", err)
		} else {
			log.Printf("Response: %s", prettyJSON.String())
		}
	}

	// Send the response
	if _, err := w.Write(responseJSON); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Handle API requests based on the API description
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	log.Printf("API: %s %s", r.Method, r.URL.Path)

	if s.apiDesc == nil {
		sendErrorResponse(w, "API description not loaded", http.StatusInternalServerError)
		return
	}

	// Find the matching endpoint
	endpoint, pathParams := s.findMatchingEndpoint(r.URL.Path)
	if endpoint == nil {
		http.NotFound(w, r)
		return
	}

	// Check if the method is supported
	methodDef, exists := endpoint.Methods[r.Method]
	if !exists {
		log.Printf("Method %s not allowed for endpoint %s", r.Method, endpoint.Path)
		sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract parameters
	queryParams := extractQueryParams(r)

	bodyParams, err := extractBodyParams(r)
	if err != nil {
		log.Printf("Warning: Failed to parse request body as JSON: %v", err)
	}

	// Prepare SQL query
	sqlQuery := ""

	// Check if SQL should be loaded from a file
	if methodDef.SQLFile != "" {
		// Determine the API description file's directory to make relative paths work
		apiDir := filepath.Dir(s.apiDescPath)

		// Build the SQL file path relative to the API description file
		sqlFilePath := filepath.Join(apiDir, methodDef.SQLFile)
		log.Printf("Loading SQL from file: %s", sqlFilePath)

		// Read the SQL file
		sqlBytes, err := os.ReadFile(sqlFilePath)
		if err != nil {
			sendErrorResponse(w, fmt.Sprintf("Failed to read SQL file: %v", err), http.StatusInternalServerError)
			return
		}

		// Use the file contents as the SQL query
		sqlQuery = string(sqlBytes)
	} else {
		// Use the inline SQL from the API definition
		sqlQuery = methodDef.SQL
	}

	// Replace named parameters with ? placeholders and build params array
	var sqlParams []any

	// If we have defined params, use them in order
	if len(methodDef.Params) > 0 {
		for _, paramName := range methodDef.Params {
			// Check path params first, then query params, then body params
			if value, ok := pathParams[paramName]; ok {
				sqlParams = append(sqlParams, value)
				sqlQuery = strings.Replace(sqlQuery, ":"+paramName, "?", 1)
			} else if value, ok := queryParams[paramName]; ok {
				sqlParams = append(sqlParams, value)
				sqlQuery = strings.Replace(sqlQuery, ":"+paramName, "?", 1)
			} else if value, ok := bodyParams[paramName]; ok {
				sqlParams = append(sqlParams, value)
				sqlQuery = strings.Replace(sqlQuery, ":"+paramName, "?", 1)
			} else {
				// Parameter not found, add nil
				sqlParams = append(sqlParams, nil)
				sqlQuery = strings.Replace(sqlQuery, ":"+paramName, "?", 1)
			}
		}
	}

	// Execute the query
	result, err := s.executeQuery(sqlQuery, sqlParams)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	s.sendJSONResponse(w, result, http.StatusOK)
}

// Handle direct SQL query requests
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	log.Printf("Query: %s", r.URL.Path)

	if r.Method != "POST" {
		sendErrorResponse(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use io.TeeReader to log the body while still allowing it to be read
	var bodyBuffer bytes.Buffer
	teeReader := io.TeeReader(r.Body, &bodyBuffer)

	// Read the body into a buffer
	_, err := io.ReadAll(teeReader)
	if err != nil {
		sendErrorResponse(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Decode the body into the QueryRequest struct
	var req QueryRequest
	if err := json.NewDecoder(&bodyBuffer).Decode(&req); err != nil {
		sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Execute the query
	result, err := s.executeQuery(req.SQL, req.Params)
	if err != nil {
		sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	s.sendJSONResponse(w, result, http.StatusOK)
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
		sendErrorResponse(w, "Invalid target URL: "+err.Error(), http.StatusBadRequest)
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

var (
	ErrReadFailed  = errors.New("read failed")
	ErrParseFailed = errors.New("parse failed")
)

// Load API description from file
func (s *Server) loadAPIDescription() (apiDesc *APIDescription, err error) {
	var data []byte
	data, err = os.ReadFile(s.apiDescPath)
	if errors.Is(os.ErrNotExist, err) {
		goto end
	}
	if err != nil {
		err = errors.Join(ErrReadFailed, err)
		goto end
	}
	apiDesc = &APIDescription{}
	err = json.Unmarshal(data, &apiDesc)
	if err != nil {
		apiDesc = nil
		err = errors.Join(ErrParseFailed, err)
		goto end
	}
end:
	return apiDesc, err
}
