package cfgstore

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func NoSuchFileOrDirectory(err error) (is bool) {
	var pathError *fs.PathError
	var errNo syscall.Errno
	if err == nil {
		goto end
	}
	if !errors.As(err, &pathError) {
		goto end
	}
	if pathError.Op != "open" {
		goto end
	}
	if !errors.As(pathError.Err, &errNo) {
		goto end
	}
	//goland:noinspection GoDirectComparisonOfErrors
	if errNo != syscall.ENOENT {
		goto end
	}
	is = true
end:
	return is
}

func ReadFileIfExists(file string) (bytes []byte, err error) {
	bytes, err = os.ReadFile(file)
	if NoSuchFileOrDirectory(err) {
		err = nil
	}
	return bytes, err
}

func GetBaseFilename(fullPath string) string {
	filename := filepath.Base(fullPath)
	return filename[:len(filename)-len(filepath.Ext(filename))]
}
