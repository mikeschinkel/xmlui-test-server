package dbpkg

import (
	"fmt"
)

var databaseMap = make(map[DatabaseType]Database, 0)

func RegisterDatabase(db Database) {

	db.SetBaseDatabase(NewBaseDatabase(db, DatabaseArgs{
		DatabaseType: db.Type(),
	}))

	databaseMap[db.Type()] = db
}

func GetRegisteredDatabase(dt DatabaseType) (db Database, err error) {
	var ok bool
	db, ok = databaseMap[dt]
	if !ok {
		err = fmt.Errorf("database type '%s' not supported", dt)
	}
	return db, err
}
