package dbpkg

import (
	"errors"
)

var ErrExtensionUnsupportedForDB = errors.New("extensions unsupported for database")

func ParseDBExtensions(dt DatabaseType, files []string) (exts []DBExtension, err error) {
	db, err := GetRegisteredDatabase(dt)
	if db == nil {
		err = errors.Join(ErrExtensionUnsupportedForDB, err)
		goto end
	}
	exts, err = db.ParseExtensions(files)
end:
	return exts, err
}
