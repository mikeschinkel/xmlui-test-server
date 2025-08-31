package xmluisvr

import (
	"errors"
	"os"

	"github.com/xmlui-org/xmluisvr/cliutil"
)

// TODO: Add functionality to set logfile with environment var or flags
const logFile = "./xmlui-local-server.log"

func RunCLI() {
	w := cliutil.NewWriter()
	logger, err := fileLogger(logFile)
	if err != nil {
		cliutil.Errorf("Failed to create and/or open log file %s: %v\n", logFile, err)
		cliutil.Errorf("Continuing without writing logs to disk\n")
	}
	err = Run(&RunArgs{
		FlagValues: parseFlagValues(),
		CLIWriter:  w,
		Logger:     logger,
	})
	switch {
	case err == nil:
		w.Printf("%s terminated gracefully", AppName)
	case errors.Is(err, ErrServerError):
		w.Errorf("%s terminated with error: %v", AppName, err)
		os.Exit(1)
	default:
		w.Errorf("Error running %s: %v", AppName, err)
		os.Exit(2)
	}
}
