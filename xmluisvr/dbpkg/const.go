package dbpkg

const ConfigSlug = "db"

const (
	DuckDBDatabase   DatabaseType = "duckdb"
	MariaDBDatabase  DatabaseType = "mariadb"
	MySQLDatabase    DatabaseType = "mysql"
	PostgresDatabase DatabaseType = "postgres"
	SQLite3Database  DatabaseType = "sqlite3"
)
