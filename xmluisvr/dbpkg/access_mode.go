package dbpkg

import "C"
import (
	"strings"

	"github.com/mattn/go-sqlite3"
)

// Constants for AccessMode values
// IMPORTANT: Number values of the AccessModes are critical to the algorithm
const (
	UnspecifiedMode AccessMode = iota
	ReadOnlyMode
	ReadWriteMode
	AdminMode
	SuperAdminMode
)

type AccessMode int

func (m AccessMode) Name() string {
	switch m {
	case SuperAdminMode:
		return "Super Admin"
	case AdminMode:
		return "Admin"
	case ReadWriteMode:
		return "Read-Write"
	case ReadOnlyMode:
		return "Read-Only"
	case UnspecifiedMode:
		fallthrough
	default:
		return "Unspecified"
	}
}

func (m AccessMode) IsRecognizedOp(op int) bool {
	_, ok := accessModeOpsDenied[op]
	return ok
}

func (m AccessMode) DeniedOp(op int, funcName string) bool {
	return !m.AllowedOp(op, funcName)
}

func (m AccessMode) AllowedOp(op int, funcName string) (allowed bool) {
	var deniedOpMode AccessMode
	var ok bool
	// Test these special cases first
	if op == sqlite3.SQLITE_FUNCTION && m < SuperAdminMode && strings.EqualFold(funcName, "load_extension") {
		goto end
	}
	deniedOpMode, ok = accessModeOpsDenied[op]
	if !ok {
		goto end
	}
	allowed = m > deniedOpMode
end:
	return allowed
}

// accessModeOpsDenied contains ops that are denied for its associated
// AccessMode, and for anything mode that has a numeric value of less than the
// access mode.
// IMPORTANT: Number values of the AccessModes are critical to the algorithm
var accessModeOpsDenied = map[int]AccessMode{
	sqlite3.SQLITE_READ:                UnspecifiedMode,
	sqlite3.SQLITE_SELECT:              UnspecifiedMode,
	sqlite3.SQLITE_ANALYZE:             UnspecifiedMode,
	sqlite3.SQLITE_INSERT:              ReadOnlyMode,
	sqlite3.SQLITE_UPDATE:              ReadOnlyMode,
	sqlite3.SQLITE_DELETE:              ReadOnlyMode,
	sqlite3.SQLITE_TRANSACTION:         ReadOnlyMode,
	sqlite3.SQLITE_FUNCTION:            ReadOnlyMode,
	sqlite3.SQLITE_COPY:                ReadOnlyMode,
	sqlite3.SQLITE_REINDEX:             ReadOnlyMode,
	sqlite3.SQLITE_SAVEPOINT:           ReadOnlyMode,
	sqlite3.SQLITE_ALTER_TABLE:         ReadWriteMode,
	sqlite3.SQLITE_CREATE_INDEX:        ReadWriteMode,
	sqlite3.SQLITE_CREATE_TABLE:        ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_INDEX:   ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_TABLE:   ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_TRIGGER: ReadWriteMode,
	sqlite3.SQLITE_CREATE_TEMP_VIEW:    ReadWriteMode,
	sqlite3.SQLITE_CREATE_TRIGGER:      ReadWriteMode,
	sqlite3.SQLITE_CREATE_VIEW:         ReadWriteMode,
	sqlite3.SQLITE_DROP_INDEX:          ReadWriteMode,
	sqlite3.SQLITE_DROP_TABLE:          ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_INDEX:     ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_TABLE:     ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_TRIGGER:   ReadWriteMode,
	sqlite3.SQLITE_DROP_TEMP_VIEW:      ReadWriteMode,
	sqlite3.SQLITE_DROP_TRIGGER:        ReadWriteMode,
	sqlite3.SQLITE_DROP_VIEW:           ReadWriteMode,
	sqlite3.SQLITE_ATTACH:              AdminMode,
	sqlite3.SQLITE_DETACH:              AdminMode,
	sqlite3.SQLITE_PRAGMA:              AdminMode,
	sqlite3.SQLITE_CREATE_VTABLE:       AdminMode,
	sqlite3.SQLITE_DROP_VTABLE:         AdminMode,
}
