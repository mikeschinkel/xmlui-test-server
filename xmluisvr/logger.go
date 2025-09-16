package xmluisvr

import (
	"bytes"
	"io"
	"log/slog"
	"os"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

var logger = common.Logger()

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
