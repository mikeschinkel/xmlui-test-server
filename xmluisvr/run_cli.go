package xmluisvr

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// TODO: Add functionality to set logfile with environment var or flags
const logFile = "./xmlui-local-server.log"

// RunCLI is the main CLI entry point for the xmlui-test-server application.
// It handles command-line argument parsing, configuration loading, and starts the server.
// This function sets up logging, loads configuration files, and delegates to Run().
//
// Exit codes:
//   - 1: Configuration loading failure
//   - 2: Server terminated with error
//   - 3: Unexpected error during server execution
//   - 4: Invalid command-line options
func RunCLI() {
	var err error
	var logger *slog.Logger
	var config *cfgldr.RootConfigV1
	var options *cfgldr.Options

	options, err = cfgldr.GetOptions()
	if err != nil {
		fprintf(os.Stderr, "Invalid option(s): %v\n", strings.Replace(err.Error(), "\n", "; ", -1))
		os.Exit(4)
	}
	writer := cliutil.NewWriter(cliutil.WriterArgs{
		Quiet:     options.Quiet,
		Verbosity: options.Verbosity,
	})
	logger, err = createFileLogger(logFile)
	if err != nil {
		writer.Errorf("Failed to create and/or open log file %s: %v\n", logFile, err)
		writer.Errorf("Continuing without writing logs to disk\n")
	}
	// TODO Incorporate loaded environment vars into GetOptions
	config, err = cfgldr.LoadRootConfigV1(common.AppConfigPath)
	if err != nil {
		writer.Errorf("Failed to load config file(s); %v\n", err)
		os.Exit(1)
	}
	// TODO: Make 10 second timeout configurable
	context.WithTimeout(context.Background(), 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = Run(ctx, &RunArgs{
		Options:   options,
		Config:    config,
		CLIWriter: writer,
		Logger:    logger,
	})
	switch {
	case err == nil:
		writer.Printf("%s terminated gracefully", common.AppName)
	case errors.Is(err, ErrServerError):
		writer.Errorf("%s terminated with error: %v", common.AppName, err)
		logger.Error("Server terminated with an error", "error", err)
		os.Exit(2)
	default:
		writer.Errorf("Error running %s: %v", common.AppName, err)
		logger.Error("Server terminated with an unexpected error", "error", err)
		os.Exit(3)
	}
}
