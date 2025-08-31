package xmluisvr

import (
	"bytes"
	"io"
	"log/slog"
	"os"
)

var logger *slog.Logger

func SetLogger(l *slog.Logger) {
	logger = l
}
func ensureLogger() {
	if logger == nil {
		panic("Must call xmluisvr.SetLogger() with a *slog.Logger before using xmluisvr package")
	}
}

func fileLogger(file string) (logger *slog.Logger, err error) {
	var w io.Writer
	w, err = os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		w = &bytes.Buffer{}
	}
	logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{}))
	return logger, err
}
