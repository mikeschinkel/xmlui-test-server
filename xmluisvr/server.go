package xmluisvr

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

// InboundProxyProtocol defines the protocol used for inbound proxy requests.
const InboundProxyProtocol = "http"

// API is a placeholder type for API functionality.
// TODO: This appears to be unused and may need removal or proper documentation.
type API struct{}

// Server represents the main HTTP server instance with all its dependencies.
// It combines database access, API configuration, HTTP routing, and logging
// into a single cohesive server implementation.
//
// The server provides these main functionalities:
//   - Static file serving from the current directory
//   - Database query execution via /query endpoint
//   - HTTP proxy functionality via /proxy endpoint
//   - Configurable API endpoints based on JSON configuration
//   - CORS middleware for cross-origin requests
type Server struct {
	db                   dbpkg.Database    // Database connection and operations
	api                  *apipkg.API       // API configuration and handlers
	options              *common.Options   // Server configuration options
	port                 common.ServerPort // HTTP server port
	sourceFile           common.Filepath   // Path to server configuration file
	mux                  *http.ServeMux    // HTTP request multiplexer
	cliutil.WriterLogger                   // Embedded logging functionality
}

// ServerArgs contains all the dependencies and configuration needed to create a Server.
type ServerArgs struct {
	Database   dbpkg.Database    // Database connection
	API        *apipkg.API       // API configuration
	Port       common.ServerPort // HTTP server port
	SourceFile common.Filepath   // Configuration file path
	Options    *common.Options   // Server options
	Writer     CLIWriter         // CLI output writer
	Logger     *slog.Logger      // Structured logger
}

// NewServer creates a new Server instance with the provided configuration.
// If no port is specified, it defaults to common.DefaultServerPort.
// The server is created with an HTTP multiplexer and embedded logging.
func NewServer(args ServerArgs) *Server {
	if args.Port == 0 {
		args.Port = common.DefaultServerPort
	}
	return &Server{
		db:           args.Database,
		api:          args.API,
		port:         args.Port,
		options:      args.Options,
		sourceFile:   args.SourceFile,
		mux:          http.NewServeMux(),
		WriterLogger: cliutil.NewWriterLogger(args.Writer, args.Logger),
	}
}

// Initialize prepares the server for operation by initializing the API,
// setting up HTTP routes, and opening the database connection.
// This method must be called before ListenAndServe().
func (s *Server) Initialize(ctx Context) (err error) {
	s.V2().InfoPrint("Initializing server")
	err = s.api.Initialize(ctx)
	if errors.Is(err, common.ErrNoAPIProvided) {
		s.Printf("No APIConfig loaded")
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
		err = s.ErrorError("Failed to opened database", "database_type", s.db.Type(), "error", err)
		goto end
	}

	s.V2().InfoPrint("Server initialized")
end:
	return err
}

// ListenAndServe starts the HTTP server and begins listening for requests.
// The server listens on the configured port and applies CORS middleware
// to all requests. This method blocks until the server shuts down or an error occurs.
func (s *Server) ListenAndServe(_ Context) (err error) {
	s.InfoLoud("Server listening", "on", s.displayHost())
	return http.ListenAndServe(s.Host(), s.corsMiddleware(s.mux))
}

// corsMiddleware applies CORS headers to all HTTP responses to enable
// cross-origin requests from web browsers. It handles preflight OPTIONS
// requests and sets permissive CORS headers.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Host returns the full host:port string for the server.
func (s *Server) Host() string {
	return fmt.Sprintf("%s:%d", common.DefaultServerHost, s.port)
}

// Port returns the server's configured port number.
func (s *Server) Port() common.ServerPort {
	return s.port
}

// addRoutes configures all HTTP routes for the server including:
//   - API endpoints (if configured)
//   - Proxy endpoint (/proxy/)
//   - Query endpoint (/query)
//   - Static file serving (/)
func (s *Server) addRoutes(ctx Context) {
	s.V2().InfoPrint("Adding HTTP server routes")
	// Handle APIConfig routes first (to match /apiFile/* before static files)
	if s.api != nil {
		apiBasePath := string(s.api.BasePath)
		if !strings.HasSuffix(apiBasePath, "/") {
			apiBasePath += "/"
		}
		route := fmt.Sprintf("GET  %s", apiBasePath)
		s.V3().Printf("  — %s\n", route)
		s.mux.HandleFunc(route, s.api.HandleAPIFunc(ctx, s.db))
	}

	// Handle proxy next
	s.V3().Printf("  — ANY  /proxy/\n")
	for _, method := range common.HTTPMethods {
		s.mux.HandleFunc(fmt.Sprintf("%s /proxy/", method), s.handleProxyFunc(method))
	}

	// Then handle query endpoint
	route := "POST /query"
	s.V3().Printf("  — %s\n", route)
	s.mux.HandleFunc(route, s.handleQueryFunc(ctx, s.db))

	route = "GET  /"
	s.V3().Printf("  — %s\n", route)
	s.mux.HandleFunc(route, s.handleRootFunc())

	s.V3().InfoPrint("HTTP server routes added")

}
