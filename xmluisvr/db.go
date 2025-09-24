package xmluisvr

import (
	// Import all database drivers to register them with the dbpkg registry.
	// This ensures all supported database types are available at runtime.
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/duckdbpkg"   // DuckDB support
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/mariadbpkg"  // MariaDB support
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/mysqlpkg"    // MySQL support
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/postgrespkg" // PostgreSQL support
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg/sqlite3pkg"  // SQLite3 support
)
