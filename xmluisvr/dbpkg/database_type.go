package dbpkg

import (
	"errors"
)

type DatabaseType string

func (dt DatabaseType) String() (s string) {
	db, err := GetRegisteredDatabase(dt)
	if err != nil {
		s = "Unspecified"
		goto end
	}
	s = db.TypeName()
end:
	return s
}

func ParseDatabaseType(connStr string) (dt DatabaseType, err error) {
	var errs []error
	for _, db := range databases {
		err := db.CheckConnection(connStr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		dt = db.Type()
		goto end
	}
	err = errors.Join(append([]error{ErrConnectStringNotSupported}, errs...)...)
end:
	return dt, err
}
