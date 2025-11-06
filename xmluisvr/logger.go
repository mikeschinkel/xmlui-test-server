package xmluisvr

import (
	"bytes"
	"io"
	"log/slog"
	"os"

	"github.com/xmlui-org/localdev/xmluisvr/common"
)

// logger is the package-level logger instance.
var logger *slog.Logger

func init() {
	common.RegisterSetLoggerFunc(func(l *slog.Logger) {
		logger = l
	})
}

// createFileLogger creates a new structured logger that writes to a file.
// If the file cannot be opened, it falls back to writing to a buffer.
// The logger uses JSON format for structured logging.
func createFileLogger(file string) (logger *slog.Logger, err error) {
	var w io.Writer
	w, err = os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		w = &bytes.Buffer{}
	}
	logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{}))
	common.SetLogger(logger)
	return logger, err
}
