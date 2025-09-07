package xmluisvr

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/xmlui-org/xmluisvr/apipkg"
	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

type RunArgs struct {
	Options   *Options
	Config    *cfgldr.RootConfigV1
	CLIWriter cliutil.Writer
	Logger    *slog.Logger
}

var (
	ErrServerError = fmt.Errorf("server terminated with an error")
)

func Run(ctx Context, args *RunArgs) (err error) {
	var server *Server
	var api *apipkg.API
	var db dbpkg.Database

	writer := args.CLIWriter

	writer.Printf("%s starting\n", common.AppName)

	err = Initialize(ctx, args)
	if err != nil {
		goto end
	}

	db, err = parseDatabase(args)
	if err != nil {
		goto end
	}

	api, err = parseAPI(args)
	if err != nil {
		goto end
	}

	// TODO LoadJSON config, apply to options
	// Initialize server
	server = NewServer(ServerArgs{
		API:      api,
		Database: db,
	})
	err = server.Initialize(ctx)
	if err != nil {
		goto end
	}
	server.showConfig()
	err = server.ListenAndServe(ctx)
	if err != nil {
		err = errors.Join(ErrServerError, err)
	}
end:
	return err
}

func Initialize(_ Context, args *RunArgs) (err error) {

	// Setting the writer allows the shorthand of being able to call cliutil.Printf()
	// and cliutil.Errorf() without having a writer injected into every func.
	cliutil.SetWriter(args.CLIWriter)

	// Setting the logger sets the package level logger variable so it is accessible
	// throughout the package.
	SetLogger(args.Logger)

	return err
}

func parseAPI(args *RunArgs) (api *apipkg.API, err error) {
	var apiCfg cfgldr.APIConfig
	var apiArgs apipkg.APIArgs

	opts := args.Options

	if opts.API != "" {
		apiCfg, err = cfgldr.LoadAPIFile(opts.API)
	}
	if apiCfg == nil {
		apiCfg = args.Config.Server.API
	}
	if apiCfg == nil {
		goto end
	}
	apiArgs, err = apipkg.MakeAPIArgs(apiCfg, args.CLIWriter, args.Logger)
	if err != nil {
		goto end
	}
	api = apipkg.NewAPI(apiArgs)
end:
	return api, err
}

func parseDatabase(args *RunArgs) (db dbpkg.Database, err error) {
	var dbArgs dbpkg.DatabaseArgs
	opts := args.Options
	dbCfg := args.Config.Database

	if opts.Database == "" {
		opts.Database = cfgldr.DefaultSqlite3Database
	}

	if dbCfg == nil {
		dbCfg = cfgldr.NewGenericDBConfig(cfgldr.GenericDBConfigArgs{
			ConnectString: opts.Database,
			Port:          opts.DatabasePort,
			Extensions:    opts.DBExtensions,
		})
	}

	dbArgs, err = dbpkg.MakeDatabaseArgs(dbCfg, args.CLIWriter, args.Logger)
	if err != nil {
		goto end
	}

	db, err = dbpkg.CreateDatabase(dbArgs)
	if err != nil {
		goto end
	}

end:
	return db, err
}
