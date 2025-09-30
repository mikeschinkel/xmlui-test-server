package cfgldr

import (
	"log/slog"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

var logger *slog.Logger

func init() {
	common.RegisterSetLoggerFunc(func(l *slog.Logger) {
		logger = l
	})
}
