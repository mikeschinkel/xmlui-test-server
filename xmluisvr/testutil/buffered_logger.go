package testutil

import (
	"log/slog"
)

var bufferedLogger *slog.Logger
var bufferedLogHandler *BufferedLogHandler

func GetBufferedLogger() (*slog.Logger, *BufferedLogHandler) {
	if bufferedLogger != nil {
		goto end
	}
	bufferedLogHandler = NewBufferedLogHandler()
	bufferedLogger = slog.New(bufferedLogHandler)
end:
	return bufferedLogger, bufferedLogHandler
}
