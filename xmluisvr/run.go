package xmluisvr

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apiutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

// Run starts the xmlui-test-server with the provided configuration and context.
// This is the main server execution function that initializes all components
// and starts the HTTP server.
//
// The function follows this execution flow:
//  1. Initialize global settings (writer, logger)
//  2. ParseBytes and validate options
//  3. Initialize database connection
//  4. Load API configuration
//  5. Create and configure server
//  6. Start HTTP server and listen for requests
//
// Returns ErrServerError if the server terminates with an error condition.
func Run(ctx Context, args *RunArgs) (err error) {
	var server *Server
	var db dbpkg.Database
	var api *apipkg.API
	var opts *common.Options

	rawOpts := args.Options

	writer := args.CLIWriter

	writer.Loud().Printf("%s starting\n", common.AppName)

	err = Initialize(ctx, args)
	if err != nil {
		goto end
	}

	opts, err = ParseOptions(args.Options)
	if err != nil {
		goto end
	}

	db, err = args.parseDatabase(ctx, opts)
	if err != nil {
		goto end
	}
	defer common.CloseOrLog(db)

	api, err = args.parseAPI(rawOpts.APIFile, db, opts)
	if err != nil {
		goto end
	}

	server, err = args.parseServer(parseServerArgs{
		db:      db,
		api:     api,
		rawOpts: rawOpts,
		opts:    opts,
		config:  args.Config.ServerConfig,
	})
	if err != nil {
		goto end
	}

	err = server.Initialize(ctx)
	if err != nil {
		goto end
	}

	server.showConfig()

	server.V2().InfoPrint("Starting server")
	err = server.ListenAndServe(ctx)
	if err != nil {
		err = errors.Join(ErrServerError, err)
	}
end:
	return err
}

// Initialize sets up global package state including the CLI writer and logger.
// This must be called before other package functions to ensure proper logging
// and output formatting.
func Initialize(_ Context, args *RunArgs) (err error) {

	// Setting the writer allows the shorthand of being able to call cliutil.Printf()
	// and cliutil.Errorf() without having a writer injected into every func.
	cliutil.SetWriter(args.CLIWriter)

	// Setting the logger sets the package level logger variable so it is accessible
	// throughout the package.
	common.SetLogger(args.Logger)

	apiutil.SetGitHubRepoURL(common.GitHubRepoURL)

	return err
}
