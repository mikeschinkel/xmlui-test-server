package xmluisvr

import (
	"log/slog"

	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

// logger is the package-level logger instance.
var logger *slog.Logger

func init() {
	common.RegisterSetLoggerFunc(func(l *slog.Logger) {
		logger = l
	})
}
