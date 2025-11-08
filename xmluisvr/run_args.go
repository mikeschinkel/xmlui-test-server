package xmluisvr

import (
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

// RunArgs contains all the configuration and dependencies needed to run the server.
// This struct is used to pass configuration from the CLI layer to the core server logic.
type RunArgs struct {
	CLIArgs []string        // Command-line arguments (currently unused)
	AppInfo appinfo.AppInfo // Developer-maintaiened application information
	Config  *Config         // Parsed configuration from files
	Options *common.Options // Parsed command-line options
}
