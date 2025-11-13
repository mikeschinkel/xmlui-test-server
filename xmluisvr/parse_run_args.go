package xmluisvr

import (
	"context"
	"log/slog"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-logutil"
	"github.com/xmlui-org/localsvr/xmluisvr/cfgldr"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

// ParseRunArgs creates a complete RunArgs from a partial RunArgs and cfgldr.Options.
// This function extracts the RunArgs construction logic from RunCLI
// so it can be reused by commands that need to call Run() directly.
//
// The input args should contain:
//   - AppInfo: Application information
//   - Config.Writer: Writer for output
//   - DirsProvider: (optional) Custom directory provider for config loading
//
// Returns a new RunArgs with:
//   - Config: Complete server configuration including logger
//   - Options: Parsed common options
func ParseRunArgs(ctx context.Context, cfgOpts *cfgldr.Options, args *RunArgs) (runArgs *RunArgs, err error) {
	var cfg *cfgldr.RootConfigV1
	var opts *common.Options
	var config *Config
	var logger *slog.Logger
	var logDir dt.DirPath
	var logFile dt.Filepath
	var writer cliutil.Writer

	writer = args.Config.Writer

	// Determine log file location using DirsProvider
	if args.DirsProvider != nil {
		// Use custom project directory (e.g., demo install dir)
		logDir, err = args.DirsProvider.ProjectDirFunc()
		if err != nil {
			goto end
		}
	} else {
		// Use current working directory
		logDir, err = dt.Getwd()
		if err != nil {
			goto end
		}
	}

	// Create log file path: <logDir>/<logPath>/<logFile>
	// e.g., ~/.config/xmlui/demos/invoice/logs/xmlui-localsvr.log
	logFile = dt.FilepathJoin3(logDir, args.AppInfo.LogPath(), args.AppInfo.LogFile())

	// Create logger
	logger, err = logutil.CreateJSONFileLogger(logFile)
	if err != nil {
		writer.Printf("Warning: Failed to create log file %s: %v\n", logFile, err)
		writer.Printf("Continuing without writing logs to disk\n")
		// Continue without logger - server will handle nil logger
		err = nil
	}

	// Load root configuration
	cfg, err = cfgldr.LoadRootConfigV1(cfgldr.LoadRootConfigV1Args{
		AppInfo:      args.AppInfo,
		Options:      cfgOpts,
		DirsProvider: args.DirsProvider,
	})
	if err != nil {
		goto end
	}

	// Parse options
	opts, err = ParseOptions(cfgOpts)
	if err != nil {
		goto end
	}

	// Parse configuration
	config, err = ParseConfig(ctx, cfg, ParseConfigArgs{
		Options: opts,
		Logger:  logger,
		Writer:  writer,
	})
	if err != nil {
		goto end
	}

	// Return new RunArgs with complete Config and Options
	runArgs = &RunArgs{
		CLIArgs:      args.CLIArgs,
		AppInfo:      args.AppInfo,
		Config:       config,
		Options:      opts,
		DirsProvider: args.DirsProvider,
	}

end:
	return runArgs, err
}
