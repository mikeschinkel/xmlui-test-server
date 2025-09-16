package xmluisvr

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// TODO: Add functionality to set logfile with environment var or flags
const logFile = "./xmlui-local-server.log"

func RunCLI() {
	var err error
	var logger *slog.Logger
	var config *cfgldr.RootConfigV1

	options := cfgldr.GetOptions()
	writer := cliutil.NewWriter()
	writer.SetQuiet(options.Quiet)
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
