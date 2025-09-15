package xmluisvr

import (
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/duckdbpkg"
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/mariadbpkg"
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/mysqlpkg"
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/postgrespkg"
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/sqlite3pkg"
)
