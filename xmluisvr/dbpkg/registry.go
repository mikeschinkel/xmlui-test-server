package dbpkg

import (
	"fmt"
)

var databases = make([]Database, 0)

func RegisterDatabase(db Database) {
	databases = append(databases, db)
}
func GetRegisteredDatabase(dt DatabaseType) (db Database, err error) {
	for _, db := range databases {
		if db.Type() != dt {
			continue
		}
		goto end
	}
	err = fmt.Errorf("database type '%s' not supported", dt)
	db = nil
end:
	return db, err
}
