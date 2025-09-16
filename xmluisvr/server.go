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
	Writer     CLIWriter
	Logger     *slog.Logger
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
		db:         args.Database,
		api:        args.API,
		port:       args.Port,
		options:    args.Options,
		sourceFile: args.SourceFile,
		mux:        http.NewServeMux(),
		Writer:     args.Writer,
		Logger:     args.Logger,
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
	cliutil.Loud().Printf("Listening on %s...\n", s.displayHost())
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
	return fmt.Sprintf("%s:%d", common.LocalHostIP, s.port)
}
func (s *Server) Port() common.ServerPort {
	return s.port
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
		s.mux.HandleFunc(method+" /proxy/", s.handleProxyFunc())
	}

	// Then handle query endpoint
	s.mux.HandleFunc("POST /query", s.handleQueryFunc(ctx, s.db))

	s.mux.HandleFunc("GET /", s.handleRootFunc())

}
