package dbpkg_test

import (
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbpkg"
)

type operation int

func TestAccessMode_Allowed(t *testing.T) {
	tests := []struct {
		m        dbpkg.AccessMode
		op       operation
		allowed  bool
		funcName string
	}{

		// Ops denied for dbpkg.UnspecifiedMode
		{op: sqlite3.SQLITE_READ, m: dbpkg.UnspecifiedMode, allowed: false},
		{op: sqlite3.SQLITE_SELECT, m: dbpkg.UnspecifiedMode, allowed: false},
		{op: sqlite3.SQLITE_ANALYZE, m: dbpkg.UnspecifiedMode, allowed: false},

		// Ops denied for dbpkg.ReadOnlyMode
		{op: sqlite3.SQLITE_INSERT, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_UPDATE, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_DELETE, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_TRANSACTION, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_COPY, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_FUNCTION, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_REINDEX, m: dbpkg.ReadOnlyMode, allowed: false},
		{op: sqlite3.SQLITE_SAVEPOINT, m: dbpkg.ReadOnlyMode, allowed: false},

		// Ops denied for dbpkg.ReadWriteMode
		{op: sqlite3.SQLITE_ALTER_TABLE, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_INDEX, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TABLE, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TEMP_INDEX, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TEMP_TABLE, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TEMP_TRIGGER, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TEMP_VIEW, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_TRIGGER, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_VIEW, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_INDEX, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TABLE, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TEMP_INDEX, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TEMP_TABLE, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TEMP_TRIGGER, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TEMP_VIEW, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_TRIGGER, m: dbpkg.ReadWriteMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_VIEW, m: dbpkg.ReadWriteMode, allowed: false},

		// Ops denied for dbpkg.AdminMode
		{op: sqlite3.SQLITE_ATTACH, m: dbpkg.AdminMode, allowed: false},
		{op: sqlite3.SQLITE_DETACH, m: dbpkg.AdminMode, allowed: false},
		{op: sqlite3.SQLITE_PRAGMA, m: dbpkg.AdminMode, allowed: false},
		{op: sqlite3.SQLITE_CREATE_VTABLE, m: dbpkg.AdminMode, allowed: false},
		{op: sqlite3.SQLITE_DROP_VTABLE, m: dbpkg.AdminMode, allowed: false},
		{op: sqlite3.SQLITE_FUNCTION, m: dbpkg.AdminMode, allowed: false, funcName: "load_extension"},

		// Ops allowed for dbpkg.ReadOnlyMode
		{op: sqlite3.SQLITE_READ, m: dbpkg.ReadOnlyMode, allowed: true},
		{op: sqlite3.SQLITE_SELECT, m: dbpkg.ReadOnlyMode, allowed: true},
		{op: sqlite3.SQLITE_ANALYZE, m: dbpkg.ReadOnlyMode, allowed: true},

		// Ops allowed for dbpkg.ReadWriteMode
		{op: sqlite3.SQLITE_INSERT, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_UPDATE, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_DELETE, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_TRANSACTION, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_COPY, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_FUNCTION, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_REINDEX, m: dbpkg.ReadWriteMode, allowed: true},
		{op: sqlite3.SQLITE_SAVEPOINT, m: dbpkg.ReadWriteMode, allowed: true},

		// Ops allowed for dbpkg.AdminMode
		{op: sqlite3.SQLITE_ALTER_TABLE, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_INDEX, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TABLE, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TEMP_INDEX, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TEMP_TABLE, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TEMP_TRIGGER, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TEMP_VIEW, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_TRIGGER, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_VIEW, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_INDEX, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TABLE, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TEMP_INDEX, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TEMP_TABLE, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TEMP_TRIGGER, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TEMP_VIEW, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_TRIGGER, m: dbpkg.AdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_VIEW, m: dbpkg.AdminMode, allowed: true},

		// Ops allowed for dbpkg.SuperAdminMode
		{op: sqlite3.SQLITE_ATTACH, m: dbpkg.SuperAdminMode, allowed: true},
		{op: sqlite3.SQLITE_DETACH, m: dbpkg.SuperAdminMode, allowed: true},
		{op: sqlite3.SQLITE_PRAGMA, m: dbpkg.SuperAdminMode, allowed: true},
		{op: sqlite3.SQLITE_CREATE_VTABLE, m: dbpkg.SuperAdminMode, allowed: true},
		{op: sqlite3.SQLITE_DROP_VTABLE, m: dbpkg.SuperAdminMode, allowed: true},
		{op: sqlite3.SQLITE_FUNCTION, m: dbpkg.SuperAdminMode, allowed: true, funcName: "load_extension"},
	}
	for _, tt := range tests {
		t.Run(tt.op.name()+"/"+tt.m.Name(), func(t *testing.T) {
			name := "DeniedOp"
			if tt.allowed {
				name = "AllowedOp"
			}
			t.Run(name, func(t *testing.T) {
				if got := tt.m.AllowedOp(int(tt.op), tt.funcName); got != tt.allowed {
					t.Errorf("AllowedOp() = %v, want %v", got, tt.allowed)
				}
			})
		})
	}
}

func (op operation) name() string {
	switch op {
	case sqlite3.SQLITE_READ:
		return "READ"
	case sqlite3.SQLITE_SELECT:
		return "SELECT"
	case sqlite3.SQLITE_INSERT:
		return "INSERT"
	case sqlite3.SQLITE_UPDATE:
		return "UPDATE"
	case sqlite3.SQLITE_DELETE:
		return "DELETE"
	case sqlite3.SQLITE_TRANSACTION:
		return "TRANSACTION"
	case sqlite3.SQLITE_ALTER_TABLE:
		return "ALTER_TABLE"
	case sqlite3.SQLITE_ANALYZE:
		return "ANALYZE"
	case sqlite3.SQLITE_CREATE_INDEX:
		return "CREATE_INDEX"
	case sqlite3.SQLITE_CREATE_TABLE:
		return "CREATE_TABLE"
	case sqlite3.SQLITE_CREATE_TEMP_INDEX:
		return "CREATE_TEMP_INDEX"
	case sqlite3.SQLITE_CREATE_TEMP_TABLE:
		return "CREATE_TEMP_TABLE"
	case sqlite3.SQLITE_CREATE_TEMP_TRIGGER:
		return "CREATE_TEMP_TRIGGER"
	case sqlite3.SQLITE_CREATE_TEMP_VIEW:
		return "CREATE_TEMP_VIEW"
	case sqlite3.SQLITE_CREATE_TRIGGER:
		return "CREATE_TRIGGER"
	case sqlite3.SQLITE_CREATE_VIEW:
		return "CREATE_VIEW"
	case sqlite3.SQLITE_DROP_INDEX:
		return "DROP_INDEX"
	case sqlite3.SQLITE_DROP_TABLE:
		return "DROP_TABLE"
	case sqlite3.SQLITE_DROP_TEMP_INDEX:
		return "DROP_TEMP_INDEX"
	case sqlite3.SQLITE_DROP_TEMP_TABLE:
		return "DROP_TEMP_TABLE"
	case sqlite3.SQLITE_DROP_TEMP_TRIGGER:
		return "DROP_TEMP_TRIGGER"
	case sqlite3.SQLITE_DROP_TEMP_VIEW:
		return "DROP_TEMP_VIEW"
	case sqlite3.SQLITE_DROP_TRIGGER:
		return "DROP_TRIGGER"
	case sqlite3.SQLITE_DROP_VIEW:
		return "DROP_VIEW"
	case sqlite3.SQLITE_ATTACH:
		return "ATTACH"
	case sqlite3.SQLITE_DETACH:
		return "DETACH"
	case sqlite3.SQLITE_PRAGMA:
		return "PRAGMA"
	case sqlite3.SQLITE_CREATE_VTABLE:
		return "CREATE_VTABLE"
	case sqlite3.SQLITE_DROP_VTABLE:
		return "DROP_VTABLE"
	case sqlite3.SQLITE_COPY:
		return "COPY"
	case sqlite3.SQLITE_FUNCTION:
		return "FUNCTION"
	case sqlite3.SQLITE_REINDEX:
		return "REINDEX"
	case sqlite3.SQLITE_SAVEPOINT:
		return "SAVEPOINT"

	}
	return "**UNSPECIFIED**"
}
