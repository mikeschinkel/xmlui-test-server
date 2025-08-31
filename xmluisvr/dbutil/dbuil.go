package dbutil

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Database interface {
	Name() string
	Type() DatabaseType
	Open() error
	Query(string, ...any) (*sql.Rows, error)
}

func CreateDatabase(dt DatabaseType, args DatabaseArgs) (db Database, err error) {
	switch dt {
	case SQLiteDatabase:
		db = NewSqliteDB(args)
	case PostgresDatabase:
		db = NewPostgresDB(args)
	default:
		return nil, fmt.Errorf("database type '%s' not supported", dt)
	}
	return db, nil
}

type db = *sql.DB
type database struct {
	db
	dbType DatabaseType
	conn   string
	writer CLIWriter
	logger *slog.Logger
}
type DatabaseArgs struct {
	DatabaseType     DatabaseType
	ConnectionString string
	ExtensionPaths   []string
	CLIWriter        CLIWriter
	Logger           *slog.Logger
}

func newDatabase(args DatabaseArgs) *database {
	return &database{
		dbType: args.DatabaseType,
		conn:   args.ConnectionString,
		writer: args.CLIWriter,
		logger: args.Logger,
	}
}

func (db *database) Name() string {
	return db.dbType.String()
}

func (db *database) Query(q string, args ...any) (*sql.Rows, error) {
	return db.db.Query(q, args...)
}
