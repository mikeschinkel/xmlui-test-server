package common

import (
	"errors"
	"io"
	"os"
)

func MustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		ensureLogger().Warn("Failed to close", "error", err)
	}
}

// CheckFileExists always returns an error indicating the status of the file. It
// is hte callers responsibility to decide which "errors" are relevant to their
// use-case.
func CheckFileExists(path string) error {
	info, err := os.Stat(path)
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

func EnsureFileExists(path string) (err error) {
	err = CheckFileExists(path)
	if errors.Is(err, ErrFileExists) {
		err = nil
	}
	return err
}
