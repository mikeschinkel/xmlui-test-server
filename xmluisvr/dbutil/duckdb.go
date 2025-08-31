package dbutil

var _ Database = (*DuckDB)(nil)

type DuckDB struct {
	*database
}

func (DuckDB) Type() DatabaseType {
	return DuckDBDatabase
}

func (d DuckDB) Open() error {
	//TODO implement me
	panic("implement me")
}

func NewDuckDB(args DatabaseArgs) *DuckDB {
	return &DuckDB{database: newDatabase(args)}
}
