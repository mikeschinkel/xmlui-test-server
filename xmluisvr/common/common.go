package common

import (
	"errors"
	"io"
	"os"
)

func CloseOrLog(c io.Closer) {
	EnsureLogger()
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("Panicked on close", "error", err)
		}
	}()
	err := c.Close()
	if err != nil {
		logger.Warn("Failed to close", "error", err)
	}
}
func LogOnError(err error) {
	EnsureLogger()
	if err != nil {
		logger.Warn("Operation failed", "error", err)
	}
}

// CheckFileExists always returns an error indicating the status of the file. It
// is hte callers responsibility to decide which "errors" are relevant to their
// use-case.
func CheckFileExists(path Filepath) error {
	info, err := os.Stat(string(path))
	if errors.Is(err, os.ErrNotExist) {
		err = errors.Join(ErrFileDoesNotExist, err)
		goto end
	}
	if err != nil {
		goto end
	}
	if info.IsDir() {
		err = errors.Join(ErrPathIsDir, err)
	}
	err = ErrFileExists
end:
	return err
}
