package dbpkg

import (
	"io"

	"github.com/xmlui-org/xmluisvr/cliutil"
)

func closeOrLog(c io.Closer) {
	if err := c.Close(); err != nil {
		cliutil.Printf("Warning: Failed to close: %v", err)
		logger.Warn("Failed to close", "error", err)
	}
}
