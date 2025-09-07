package dbpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/fsutil"
)

type Database interface {
	TypeName() string
	ConnectString() string
	Type() DatabaseType
	Open() error
	Initialize(ctx Context) error
	Query(Context, string, ...any) (*sql.Rows, error)
	CheckConnection(string) error
	ParseExtensions(files []string) ([]DBExtension, error)
	CreateNew(args DatabaseArgs) Database
	Extensions() []DBExtension
	fmt.Stringer
}

func MakeDatabaseArgs(cfg cfgldr.DatabaseConfig, w cliutil.Writer, l *slog.Logger) (dbArgs DatabaseArgs, err error) {
	var dt DatabaseType
	var exts []DBExtension

	dt, err = ParseDatabaseType(cfg.ConnectString())
	if err != nil {
		goto end
	}
	exts, err = ParseDBExtensions(dt, cfg.Extensions())
	if err != nil {
		goto end
	}
	dbArgs = DatabaseArgs{
		DatabaseType:  dt,
		ConnectString: cfg.ConnectString(),
		Port:          cfg.Port(),
		Extensions:    exts,
		CLIWriter:     w,
		Logger:        l,
	}
end:
	return dbArgs, err
}

func CreateDatabase(args DatabaseArgs) (db Database, err error) {
	var ndb Database
	db, err = GetRegisteredDatabase(args.DatabaseType)
	if err != nil {
		goto end
	}
	ndb = db.CreateNew(args)
end:
	return ndb, err
}

type DBExtension interface {
	Name() string
	DBExtension()
}

type BaseDatabase struct {
	*sql.DB
	dbType      DatabaseType
	conn        string
	writer      CLIWriter
	logger      *slog.Logger
	parent      Database
	initialized bool
}
type DatabaseArgs struct {
	DatabaseType  DatabaseType
	ConnectString string
	Port          int
	Extensions    []DBExtension
	CLIWriter     CLIWriter
	Logger        *slog.Logger
}

func NewBaseDatabase(parent Database, args DatabaseArgs) *BaseDatabase {
	return &BaseDatabase{
		parent: parent,
		dbType: args.DatabaseType,
		conn:   args.ConnectString,
		writer: args.CLIWriter,
		logger: args.Logger,
	}
}

func (db *BaseDatabase) Initialize(_ Context) (err error) {
	if db.initialized {
		goto end
	}
	err = db.parent.Open()
	db.initialized = true
end:
	return err
}

func (db *BaseDatabase) ConnectString() string {
	return db.conn
}

func (db *BaseDatabase) ParseExtensions(_ []string) ([]DBExtension, error) {
	// Stub for those databases for which we do not current support extensions.
	return nil, nil
}

func (db *BaseDatabase) Query(ctx Context, q string, params ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, q, params...)
}

func (db *BaseDatabase) CheckConnection(cs string) (err error) {
	conn, err := sql.Open(string(db.dbType), cs)
	if err != nil {
		goto end
	}
	common.MustClose(conn)
end:
	return err
}

func (db *BaseDatabase) CheckFileConnection(parent Database, cs string) (err error) {
	err = common.CheckFileExists(cs)
	switch {
	case errors.Is(err, common.ErrFileDoesNotExist):
		goto end
	case errors.Is(err, common.ErrPathIsDir):
		goto end
	}
	err = parent.CheckConnection(cs)
end:
	return err
}

func (db *BaseDatabase) HomeRelativeFile() string {
	absPath, err := filepath.Abs(db.conn)
	if err != nil {
		panic(fmt.Sprintf("Failed to get absolute path of '%s': %v", db.conn, err))
	}
	return fsutil.HomeRelative(absPath)
}

func (db *BaseDatabase) Extensions() []DBExtension {
	return make([]DBExtension, 0)
}
