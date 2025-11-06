package sqlite3pkg

import "C"
import (
	"github.com/mattn/go-sqlite3"
	"github.com/xmlui-org/localdev/xmluisvr/dbpkg"
)

func isRecognizedOp(op int) bool {
	_, ok := accessModeOpsDenied[op]
	return ok
}

// accessModeOpsDenied maps SQLite operations to the highest access mode at which
// they are DENIED. Operations are allowed when the current access mode is GREATER
// than the denied mode.
//
// # Authorization Algorithm
//
// The check is: allowed = currentAccessMode > accessModeOpsDenied[operation]
//
// Access mode hierarchy (from dbpkg/access_mode.go):
//   - UnspecifiedAccessMode = 0  (sentinel, auto-converted to default in production)
//   - ReadOnlyMode          = 1
//   - ReadWriteMode         = 2
//   - AdminMode             = 3
//   - SuperAdminMode        = 4
//
// Example: SQLITE_READ is mapped to UnspecifiedAccessMode (0)
//   - At ReadOnlyMode (1): 1 > 0 = true  → ALLOWED
//   - At ReadWriteMode+:   always true   → ALLOWED
//
// Example: SQLITE_INSERT is mapped to ReadOnlyMode (1)
//   - At ReadOnlyMode (1): 1 > 1 = false → DENIED
//   - At ReadWriteMode (2): 2 > 1 = true → ALLOWED
//
// IMPORTANT: The numeric values of AccessModes are critical to this algorithm.
// Changing them will break the authorization logic.
//
// See: sqlite3.go IsAuthorizedSQLite3Operation() for the authorization implementation
// See: access_mode_test.go for comprehensive test coverage
var accessModeOpsDenied = map[int]dbpkg.AccessMode{
	sqlite3.SQLITE_READ:                dbpkg.UnspecifiedAccessMode,
	sqlite3.SQLITE_SELECT:              dbpkg.UnspecifiedAccessMode,
	sqlite3.SQLITE_ANALYZE:             dbpkg.UnspecifiedAccessMode,
	sqlite3.SQLITE_INSERT:              dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_UPDATE:              dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_DELETE:              dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_TRANSACTION:         dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_FUNCTION:            dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_COPY:                dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_REINDEX:             dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_SAVEPOINT:           dbpkg.ReadOnlyMode,
	sqlite3.SQLITE_ALTER_TABLE:         dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_INDEX:        dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TABLE:        dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_INDEX:   dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_TABLE:   dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_TRIGGER: dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_VIEW:    dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_TRIGGER:      dbpkg.ReadWriteMode,
	sqlite3.SQLITE_CREATE_VIEW:         dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_INDEX:          dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TABLE:          dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_INDEX:     dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_TABLE:     dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_TRIGGER:   dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_VIEW:      dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_TRIGGER:        dbpkg.ReadWriteMode,
	sqlite3.SQLITE_DROP_VIEW:           dbpkg.ReadWriteMode,
	sqlite3.SQLITE_ATTACH:              dbpkg.AdminMode,
	sqlite3.SQLITE_DETACH:              dbpkg.AdminMode,
	sqlite3.SQLITE_PRAGMA:              dbpkg.AdminMode,
	sqlite3.SQLITE_CREATE_VTABLE:       dbpkg.AdminMode,
	sqlite3.SQLITE_DROP_VTABLE:         dbpkg.AdminMode,
}
