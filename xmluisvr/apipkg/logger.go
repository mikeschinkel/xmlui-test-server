package apipkg

import (
	"log/slog"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
)

// logger provides package-level logging functionality using the common logger instance.
var logger *slog.Logger

func init() {
	common.RegisterSetLoggerFunc(func(l *slog.Logger) {
		logger = l
		rfc9457.SetLogger(l)
	})
}
