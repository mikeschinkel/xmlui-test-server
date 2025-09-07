package xmluisvr

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
)

// TODO: Add functionality to set logfile with environment var or flags
const logFile = "./xmlui-local-server.log"

func RunCLI() {
	writer := cliutil.NewWriter()
	logger, err := fileLogger(logFile)
	if err != nil {
		writer.Errorf("Failed to create and/or open log file %s: %v\n", logFile, err)
		writer.Errorf("Continuing without writing logs to disk\n")
	}
	options := parseOptions()
	config, err := cfgldr.LoadRootConfigV1(common.AppConfigPath)
	// TODO Incorporate loaded config and environment vars here, somehow
	context.WithTimeout(context.Background(), 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = Run(ctx, &RunArgs{
		Config:    config,
		Options:   options,
		CLIWriter: writer,
		Logger:    logger,
	})
	switch {
	case err == nil:
		writer.Printf("%s terminated gracefully", common.AppName)
	case errors.Is(err, ErrServerError):
		writer.Errorf("%s terminated with error: %v", common.AppName, err)
		logger.Error("Server terminated with an error", "error", err)
		os.Exit(1)
	default:
		writer.Errorf("Error running %s: %v", common.AppName, err)
		logger.Error("Server terminated with an unexpected error", "error", err)
		os.Exit(2)
	}
}
