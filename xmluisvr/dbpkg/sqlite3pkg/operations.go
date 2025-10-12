package sqlite3pkg

import "C"
import (
	"github.com/mattn/go-sqlite3"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

func isRecognizedOp(op int) bool {
	_, ok := accessModeOpsDenied[op]
	return ok
}

// accessModeOpsDenied contains ops that are denied for its associated
// AccessMode, and for anything mode that has a numeric value of less than the
// access mode.
// IMPORTANT: Number values of the AccessModes are critical to the algorithm
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
	sqlite3.SQLITE_PRAGMA:              dbpkg.ReadOnlyMode,
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
	sqlite3.SQLITE_CREATE_VTABLE:       dbpkg.AdminMode,
	sqlite3.SQLITE_DROP_VTABLE:         dbpkg.AdminMode,
}
