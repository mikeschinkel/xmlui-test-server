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

const InboundProxyProtocol = "http"

type API struct{}

type Server struct {
	db         dbpkg.Database
	api        *apipkg.API
	options    *common.Options
	port       common.ServerPort
	sourceFile common.Filepath
	mux        *http.ServeMux
	cliutil.WriterLogger
}

type ServerArgs struct {
	Database   dbpkg.Database
	API        *apipkg.API
	Port       common.ServerPort
	SourceFile common.Filepath
	Options    *common.Options
	Writer     CLIWriter
	Logger     *slog.Logger
}

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

func (s *Server) ListenAndServe(_ Context) (err error) {
	s.InfoLoud("Server listening", "on", s.displayHost())
	return http.ListenAndServe(s.Host(), s.corsMiddleware(s.mux))
}

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

func (s *Server) Host() string {
	return fmt.Sprintf("%s:%d", common.DefaultServerHost, s.port)
}
func (s *Server) Port() common.ServerPort {
	return s.port
}

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
