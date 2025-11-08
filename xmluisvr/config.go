package xmluisvr

import (
	"log/slog"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/xmlui-org/localsvr/xmluisvr/apipkg"
	"github.com/xmlui-org/localsvr/xmluisvr/dbpkg"
)

type Config struct {
	Server   *Server
	Database dbpkg.Database
	Logger   *slog.Logger
	Writer   cliutil.Writer
}

func (c Config) API() (api *apipkg.API) {
	if c.Server == nil {
		goto end
	}
	api = c.Server.API
end:
	return api
}

type ConfigArgs struct {
	Server   *Server
	Database dbpkg.Database
	Logger   *slog.Logger
	Writer   cliutil.Writer
}

func NewConfig(args ConfigArgs) *Config {
	return &Config{
		Server:   args.Server,
		Database: args.Database,
		Logger:   args.Logger,
		Writer:   args.Writer,
	}
}
