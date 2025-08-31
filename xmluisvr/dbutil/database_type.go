package dbutil

const (
	SQLiteDatabase   DatabaseType = "sqlite"
	PostgresDatabase DatabaseType = "postgres"
	DuckDBDatabase   DatabaseType = "duckdb"
)

type DatabaseType string

func (dt DatabaseType) String() string {
	switch dt {
	case SQLiteDatabase:
		return "SQLite"
	case PostgresDatabase:
		return "PostgreSQL"
	case DuckDBDatabase:
		return "DuckDB"
	}
	return "Unspecified"
}
