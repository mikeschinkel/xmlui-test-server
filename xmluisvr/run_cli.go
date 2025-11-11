package xmluisvr

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/xmlui-org/localsvr/xmluisvr/cfgldr"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

// TODO: Add functionality to set logfile with environment var or flags
const logFile = "./xmlui-local-server.log"

// RunCLI is the main CLI entry point for the xmlui-test-server application.
// It handles command-line argument parsing, configuration loading, and starts the server.
// This function sets up logging, loads configuration files, and delegates to Run().
//
// If cfgOpts is nil, it will parse command-line flags itself (standalone mode).
// If cfgOpts is provided, it will use those options (composed mode with xmlui CLI).
//
// Exit codes:
//   - 1: Options parsing failure
//   - 2: Configuration loading failure
//   - 3: Configuration parsing failure
//   - 4: Known runtime error
//   - 5: Unknown runtime error
func RunCLI(cfgOpts *cfgldr.Options) {
	var err error
	var logger *slog.Logger
	var cfg *cfgldr.RootConfigV1
	var opts *common.Options
	var config *Config
	var wl cliutil.WriterLogger

	if cfgOpts == nil {
		cfgOpts, err = cfgldr.GetOptions()
		if err != nil {
			fprintf(os.Stderr, "Invalid option(s): %v\n", strings.Replace(err.Error(), "\n", "; ", -1))
			os.Exit(cliutil.ExitOptionsParseError)
		}
	}

	//goland:noinspection GoDfaErrorMayBeNotNil
	writer := cliutil.NewWriter(&cliutil.WriterArgs{
		Quiet:     cfgOpts.Quiet,
		Verbosity: cliutil.Verbosity(cfgOpts.Verbosity),
	})
	logger, err = createFileLogger(logFile)
	if err != nil {
		writer.Errorf(
			"Failed to create and/or open log file %s: %v\n",
			"Continuing without writing logs to disk",
			logFile, err,
		)
	}
	wl = cliutil.NewWriterLogger(writer, logger)

	cfg, err = cfgldr.LoadRootConfigV1(cfgldr.LoadRootConfigV1Args{
		AppInfo: AppInfo(),
		Options: cfgOpts,
	})
	if err != nil {
		writer.Errorf("Failed to load config file(s); %v\n", err)
		os.Exit(cliutil.ExitConfigLoadError)
	}

	// TODO Incorporate loaded environment vars into GetOptions
	opts, err = ParseOptions(cfgOpts)
	if err != nil {
		fprintf(os.Stderr, "Failed while parsing options: %v\n", err)
		os.Exit(cliutil.ExitOptionsParseError)
	}

	// TODO: Make 10 second timeout configurable
	context.WithTimeout(context.Background(), 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err = ParseConfig(ctx, cfg, ParseConfigArgs{
		Options:      opts,
		Logger:       logger,
		Writer:       writer,
		DirsProvider: nil, // This should be nil, use a provider only for testing
	})
	if err != nil {
		_ = wl.ErrorError("Failed to parse configuration", "error", err)
		os.Exit(cliutil.ExitConfigParseError)
	}
	defer common.CloseOrLog(config.Database)

	err = Run(ctx, &RunArgs{
		CLIArgs: nil,
		AppInfo: AppInfo(),
		Config:  config,
		Options: opts,
	})

	switch {
	case err == nil:
		writer.Printf("%s terminated gracefully", common.AppName)
	case errors.Is(err, ErrServerError):
		_ = wl.ErrorError("CLI terminated with error:",
			"cli_name", common.AppName,
			"exe_name", common.ExeName,
			"error", err,
		)
		os.Exit(cliutil.ExitKnownRuntimeError)
	default:
		_ = wl.ErrorError("Server terminated with an unexpected error", err)
		os.Exit(cliutil.ExitUnknownRuntimeError)
	}
}
