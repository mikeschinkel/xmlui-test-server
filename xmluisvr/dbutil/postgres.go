package dbutil

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var _ Database = (*PostgresDB)(nil)

type PostgresDB struct {
	*database
}

func (*PostgresDB) Type() DatabaseType {
	return PostgresDatabase
}

func NewPostgresDB(args DatabaseArgs) *PostgresDB {
	return &PostgresDB{database: newDatabase(args)}
}

func (p *PostgresDB) Open() (err error) {
	// Using PostgreSQL i
	p.writer.Printf("Using PostgreSQL database\n")
	p.db, err = sql.Open("postgres", p.conn)
	if err != nil {
		err = errors.Join(ErrConnFailed, err)
	}
	return err
}
