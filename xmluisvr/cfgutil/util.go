package cfgutil

import (
	"io"
)

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		logger.Error(err.Error())
	}
}
