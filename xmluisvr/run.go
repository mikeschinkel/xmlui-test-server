package xmluisvr

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/xmlui-org/xmluisvr/cliutil"
)

type RunArgs struct {
	FlagValues *FlagValues
	CLIWriter  cliutil.Writer
	Logger     *slog.Logger
}

var ErrServerError = fmt.Errorf("server terminated with an error")

func Run(args *RunArgs) (err error) {
	var flags *FlagValues
	var server *Server
	writer := args.CLIWriter

	writer.Printf("%s starting\n", AppName)

	err = Initialize(args)
	if err != nil {
		goto end
	}

	flags = args.FlagValues
	// Initialize server
	server = NewServer(ServerArgs{
		DatabaseType:     flags.DatabaseType(),
		ConnectionString: flags.ConnectionString(),
		ExtensionPaths:   flags.ExtensionPaths(),
		APIPath:          flags.ApiDesc,
		Verbose:          flags.Verbose,
	})
	err = server.Initialize()
	if err != nil {
		goto end
	}
	server.showConfig()
	err = server.ListenAndServe()
	if err != nil {
		err = errors.Join(ErrServerError, err)
	}
end:
	return err
}

func Initialize(args *RunArgs) (err error) {

	// Setting the writer allows the shorthand of being able to call cliutil.Printf()
	// and cliutil.Errorf() without having a writer injected into every func.
	cliutil.SetWriter(args.CLIWriter)

	// Setting the logger sets the package level logger variable so it is accessible
	// throughout the package.
	SetLogger(args.Logger)

	return err
}
