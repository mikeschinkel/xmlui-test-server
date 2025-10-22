package xmluisvr

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"

	. "github.com/xmlui-org/xmlui-test-server/xmluisvr/doterr"
)

// RunArgs contains all the configuration and dependencies needed to run the server.
// This struct is used to pass configuration from the CLI layer to the core server logic.
type RunArgs struct {
	CLIArgs   []string             // Command-line arguments (currently unused)
	Options   *cfgldr.Options      // Parsed command-line options
	Config    *cfgldr.RootConfigV1 // Loaded configuration from files
	CLIWriter cliutil.Writer       // Writer for CLI output and logging
	Logger    *slog.Logger         // Structured logger instance
}

// ParseOptions converts raw options from cfgldr.Options into
// validated common.Options. This method performs validation and type conversion
// for all XMLUI Test Server options.
func ParseOptions(rawOpts *cfgldr.Options) (opts *common.Options, err error) {
	var errs []error
	opts = &common.Options{
		AllowUntrustedQueries: rawOpts.AllowUntrustedQueries,
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
	opts.DBBootstrapFile, err = common.ParseFilepath(rawOpts.DBBootstrapFile)
	errs = append(errs, err)
	opts.Verbosity, err = common.ParseVerbosity(rawOpts.Verbosity)
	errs = append(errs, err)
	opts.ErrorStyle, err = common.ParseErrorStyle(rawOpts.ErrorStype)
	errs = append(errs, err)

	return opts, CombineErrs(errs)
}

// parseAPI loads and creates the API configuration from either a file or the root config.
// If no API file is specified or found, it falls back to the API configuration
// embedded in the root configuration.
func (args *RunArgs) parseAPI(apiFile string, db dbpkg.Database, opts *common.Options) (api *apipkg.API, err error) {
	var apiCfg cfgldr.APIConfig

	apiCfg, err = cfgldr.LoadAPIFileIfExists(apiFile)
	if err != nil {
		goto end
	}
	if apiCfg.IsNil() {
		apiCfg = args.Config.APIConfig()
	}
	api, err = apipkg.CreateAPI(apipkg.CreateAPIArgs{
		Database: db,
		Config:   apiCfg,
		Options:  opts,
		Writer:   args.CLIWriter,
		Logger:   args.Logger,
	})
end:
	return api, err
}

// parseDatabase initializes the database connection using the provided configuration and options.
// It sets up the database with the specified connection string and any configured extensions.
func (args *RunArgs) parseDatabase(ctx Context, opts *common.Options) (db dbpkg.Database, err error) {
	dbCfg := args.Config.DBConfig

	if opts.ConnectString == "" {
		opts.ConnectString = common.ConnectString(cfgldr.DefaultSQLite3Database)
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

// parseServerArgs contains the dependencies needed to create a Server instance.
type parseServerArgs struct {
	db      dbpkg.Database      // Database connection
	api     *apipkg.API         // API configuration and handlers
	rawOpts *cfgldr.Options     // Raw command-line options
	opts    *common.Options     // Parsed and validated options
	config  cfgldr.ServerConfig // Server configuration
}

// parseServer creates and configures a new Server instance with all dependencies.
// It validates the HTTP port and creates the server with the provided database,
// API configuration, and other settings.
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
