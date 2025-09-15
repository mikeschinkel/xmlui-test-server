package xmluisvr

import (
	"errors"
	"fmt"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

var (
	ErrServerError = fmt.Errorf("server terminated with an error")
)

func Run(ctx Context, args *RunArgs) (err error) {
	var server *Server
	var db dbpkg.Database
	var api *apipkg.API
	var opts *common.Options

	rawOpts := args.Options

	writer := args.CLIWriter

	writer.Printf("%s starting\n", common.AppName)

	err = Initialize(ctx, args)
	if err != nil {
		goto end
	}

	opts, err = args.parseOptions()
	if err != nil {
		goto end
	}

	db, err = args.parseDatabase(ctx, opts)
	if err != nil {
		goto end
	}

	api, err = args.parseAPI(rawOpts.APIFile)
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
	common.SetLogger(args.Logger)

	return err
}
