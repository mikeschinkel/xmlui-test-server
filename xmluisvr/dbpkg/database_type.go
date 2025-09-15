package dbpkg

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type DatabaseType string

//func (dt DatabaseType) String() (s string) {
//	db, err := GetRegisteredDatabase(dt)
//	if err != nil {
//		s = "Unspecified"
//		goto end
//	}
//	s = db.TypeName()
//end:
//	return s
//}

func ParseDatabaseType(ctx Context, connStr string) (dt DatabaseType, err error) {
	var errs []error
	var cs common.ConnectString
	for dbType, db := range databaseMap {
		cs, err = db.ParseConnectString(connStr)
		if err != nil {
			goto end
		}
		err := db.CheckConnection(ctx, dbType, cs)
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
