package duckdbpkg

import (
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

const DuckDBDatabase dbpkg.DatabaseType = "duckdb"

func init() {
	dbpkg.RegisterDatabase(&DuckDB{})
}

var _ dbpkg.Database = (*DuckDB)(nil)

type database = dbpkg.BaseDatabase

type DuckDB struct {
	*database
}

func (d *DuckDB) String() string {
	return d.HomeRelativeFile()
}

func (d *DuckDB) TypeName() string {
	return "DuckDB"
}

func (d *DuckDB) CheckConnection(cs string) (err error) {
	return d.CheckFileConnection(d, cs)
}

func (*DuckDB) Type() dbpkg.DatabaseType {
	return DuckDBDatabase
}

func (d *DuckDB) Open() error {
	//TODO implement me
	panic("implement me")
}

func NewDuckDB(args dbpkg.DatabaseArgs) *DuckDB {
	db := &DuckDB{}
	db.database = dbpkg.NewBaseDatabase(db, args)
	return db
}
func (*DuckDB) CreateNew(args dbpkg.DatabaseArgs) dbpkg.Database {
	return NewDuckDB(args)
}
