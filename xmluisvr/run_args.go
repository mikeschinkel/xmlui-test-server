package xmluisvr

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

type RunArgs struct {
	CLIArgs   []string
	Options   *cfgldr.Options
	Config    *cfgldr.RootConfigV1
	CLIWriter cliutil.Writer
	Logger    *slog.Logger
}

func (args *RunArgs) parseOptions() (opts *common.Options, err error) {
	var errs []error
	rawOpts := args.Options

	opts = &common.Options{
		Quiet: rawOpts.Quiet,
	}
	opts.Timeout, err = common.ParseTimeDurationEx(strconv.Itoa(rawOpts.Timeout))
	errs = append(errs, err)
	opts.HTTPPort, err = common.ParseServerPort(rawOpts.HTTPPort, common.ZeroOk)
	errs = append(errs, err)
	opts.DBExtensionFiles, err = common.ParseFilepaths(rawOpts.DBExtensionFiles)
	errs = append(errs, err)
	opts.APIFile, err = common.ParseFilepath(rawOpts.APIFile)
	errs = append(errs, err)
	opts.ConnectString, err = common.ParseConnectString(rawOpts.ConnectString)
	errs = append(errs, err)
	opts.DBPort, err = common.ParseServerPort(rawOpts.DBPort, common.ZeroOk)
	errs = append(errs, err)
	opts.DBSchemaFile, err = common.ParseFilepath(rawOpts.DBSchemaFile)
	errs = append(errs, err)

	return opts, errors.Join(errs...)
}

func (args *RunArgs) parseAPI(apiFile string) (api *apipkg.API, err error) {
	var apiCfg cfgldr.APIConfig

	apiCfg, err = cfgldr.LoadAPIFileIfExists(apiFile)
	if err != nil {
		goto end
	}
	if apiCfg.IsNil() {
		apiCfg = args.Config.APIConfig()
	}
	api, err = apipkg.CreateAPI(apipkg.CreateAPIArgs{
		Config: apiCfg,
		Writer: args.CLIWriter,
		Logger: args.Logger,
	})
end:
	return api, err
}

func (args *RunArgs) parseDatabase(ctx Context, opts *common.Options) (db dbpkg.Database, err error) {
	dbCfg := args.Config.DBConfig

	if opts.ConnectString == "" {
		opts.ConnectString = cfgldr.DefaultSQLite3Database
	}

	db, err = dbpkg.ParseDatabase(ctx, dbCfg, dbpkg.ParseDatabaseArgs{
		Options: opts,
		Writer:  args.CLIWriter,
		Logger:  args.Logger,
	})
	if err != nil {
		goto end
	}

end:
	return db, err
}

type parseServerArgs struct {
	db      dbpkg.Database
	api     *apipkg.API
	rawOpts *cfgldr.Options
	opts    *common.Options
	config  cfgldr.ServerConfig
}

func (args *RunArgs) parseServer(sArgs parseServerArgs) (svr *Server, err error) {
	var sourceFile common.Filepath
	var svrCfg *cfgldr.ServerConfigV1
	var ok bool

	opts := sArgs.opts

	if args.Options.HTTPPort == 0 {
		args.Options.HTTPPort = common.DefaultServerPort
	}
	opts.HTTPPort, err = common.ParseServerPort(args.Options.HTTPPort, common.ZeroInvalid)
	if err != nil {
		goto end
	}

	svrCfg, ok = sArgs.config.(*cfgldr.ServerConfigV1)
	if !ok {
		panic(fmt.Sprintf("Failed to type assert value of type %T to type %T",
			sArgs.config,
			(*cfgldr.ServerConfigV1)(nil),
		))
	}
	sourceFile, err = common.ParseFilepath(svrCfg.SourceFile)
	svr = NewServer(ServerArgs{
		Database:   sArgs.db,
		API:        sArgs.api,
		Port:       opts.HTTPPort,
		SourceFile: sourceFile,
		Options:    opts,
		Writer:     args.CLIWriter,
		Logger:     args.Logger,
	})
end:
	return svr, err
}
