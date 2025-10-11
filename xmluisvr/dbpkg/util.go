package dbpkg

import (
	"io"
	"log"
)

func closeOrLog(c io.Closer) {
	// TODO: Change log to using app's *slog.Logger
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panicked on close: %v", err)
		}
	}()
	err := c.Close()
	if err != nil {
		log.Printf("Failed to close: %v", err)
	}
}
