# sqlite3pkg — XML LocalDev server-specific Sqlite3 package

SQLite3 database implementation for xmlui-test-server with operation-level access control.

## Access Mode Authorization

This package implements a fine-grained authorization system that controls which SQLite operations can be executed based on the configured access mode.

### Access Mode Hierarchy

The system uses five numeric access modes, where higher numbers grant more permissions:

| Mode | Value | Description |
|------|-------|-------------|
| `UnspecifiedAccessMode` | 0 | Sentinel value only - auto-converted to `ReadWriteMode` at runtime |
| `ReadOnlyMode` | 1 | Allows read operations: `SELECT`, `READ`, `ANALYZE` |
| `ReadWriteMode` | 2 | Allows data modification: `INSERT`, `UPDATE`, `DELETE`, `TRANSACTION` |
| `AdminMode` | 3 | Allows schema changes: `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE` |
| `SuperAdminMode` | 4 | Allows all operations including `PRAGMA`, `ATTACH`, `DETACH` |

### How Authorization Works

The authorization algorithm uses a **"denied at or below"** model:

1. Each SQLite operation is mapped to the **highest access mode at which it is still DENIED**
2. An operation is allowed when: `currentAccessMode > deniedAtMode`

**Example 1: Read Operations**

`SQLITE_READ` is mapped to `UnspecifiedAccessMode` (0):
- At `UnspecifiedAccessMode` (0): `0 > 0 = false` → **DENIED**
- At `ReadOnlyMode` (1): `1 > 0 = true` → **ALLOWED**
- At `ReadWriteMode` (2): `2 > 0 = true` → **ALLOWED**
- At all higher modes → **ALLOWED**

**Example 2: Write Operations**

`SQLITE_INSERT` is mapped to `ReadOnlyMode` (1):
- At `ReadOnlyMode` (1): `1 > 1 = false` → **DENIED**
- At `ReadWriteMode` (2): `2 > 1 = true` → **ALLOWED**
- At all higher modes → **ALLOWED**

**Example 3: Admin Operations**

`SQLITE_PRAGMA` is mapped to `AdminMode` (3):
- At `AdminMode` (3): `3 > 3 = false` → **DENIED**
- At `SuperAdminMode` (4): `4 > 3 = true` → **ALLOWED**

### Implementation

The authorization check is performed in `IsAuthorizedSQLite3Operation()` in `sqlite3.go`:

```go
func (s *SQLite3) IsAuthorizedSQLite3Operation(op int, funcName string) bool {
    deniedOpMode, ok := accessModeOpsDenied[op]
    if !ok {
        return false  // Unknown operations are denied
    }
    return s.database.AccessMode > deniedOpMode
}
```

The complete mapping of operations to denied modes is defined in `operations.go`. See that file for the full list.

### Configuration

Access mode can be configured via:

1. **Config file** (`test-server.json`):
   ```json
   {
     "database": {
       "type": "sqlite3",
       "access_mode": 2
     }
   }
   ```

2. **Programmatically**:
   ```go
   db := sqlite3pkg.NewSQLite3(sqlite3pkg.SQLite3Args{
       DatabaseArgs: dbpkg.DatabaseArgs{
           AccessMode: dbpkg.ReadWriteMode,
       },
   })
   ```

If `AccessMode` is not specified (or set to `UnspecifiedAccessMode`), it defaults to `ReadWriteMode`.

### Testing

See `access_mode_test.go` for comprehensive test coverage of all access modes and operations.

### Important Notes

⚠️ **The numeric values of access modes are critical to the algorithm.** Changing them will break authorization logic.

⚠️ **`AccessMode` is stored in `BaseDatabase`, not in the `SQLite3` struct.** The SQLite3-specific field was removed to avoid duplication and potential footguns.

⚠️ **`UnspecifiedAccessMode` never exists at runtime.** It's automatically converted to a sensible default (`ReadWriteMode`) by `NewBaseDatabase()`.

## Links
- [Steampipe Github Extension](https://steampipe.io/docs/steampipe_sqlite/configure)
  - [Docs](https://hub.steampipe.io/plugins/turbot/github)
  - [Github Repo](https://github.com/turbot/steampipe-plugin-github/tree/main)
- [Possible Solutions to Concurrency Problems — Github Issues](https://github.com/mattn/go-sqlite3/issues/1179)