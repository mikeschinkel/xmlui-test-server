package dbpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/fsutil"
)

type BaseDatabase struct {
	*sql.DB
	dbType        DatabaseType
	conn          string
	Writer        CLIWriter
	Logger        *slog.Logger
	parent        Database
	SchemaQueries *MultipartQuery
	OnOpenQueries *MultipartQuery
	extensions    []DBExtension
	sourceFile    common.Filepath
	options       *common.Options
	AccessMode    AccessMode
	Initialized   bool
}

func NewBaseDatabase(parent Database, args DatabaseArgs) *BaseDatabase {
	if args.AccessMode == UnspecifiedMode {
		args.AccessMode = ReadWriteMode
	}
	return &BaseDatabase{
		parent:        parent,
		dbType:        args.DatabaseType,
		conn:          args.ConnectString,
		SchemaQueries: args.SchemaQueries,
		OnOpenQueries: args.OnOpenQueries,
		extensions:    args.Extensions,
		options:       args.Options,
		AccessMode:    args.AccessMode,
		sourceFile:    args.SourceFile,
		Writer:        args.CLIWriter,
		Logger:        args.Logger,
	}
}

func (db *BaseDatabase) SourceFile() common.Filepath {
	return db.sourceFile
}

func (db *BaseDatabase) QueryFileExt() string {
	return ".sql"
}

func (db *BaseDatabase) HasExtensions() bool {
	return len(db.extensions) > 0
}

func (db *BaseDatabase) LoadExtensions() error {
	var errs = make([]error, 0)

	for _, ext := range db.Extensions() {
		errs = append(errs, db.parent.LoadExtension(ext))
	}
	return errors.Join(errs...)
}

func (db *BaseDatabase) LoadExtension(_ DBExtension) error {
	db.checkForExtensions()
	// Stub for those databases for which we do not current support extensions.
	return errors.Join(ErrExtensionsUnsupportedForDBType, fmt.Errorf("database_type=%s", db.dbType))
}

func (db *BaseDatabase) ParseExtension(_ DBExtensionConfig) (DBExtension, error) {
	db.checkForExtensions()
	// Stub for those databases for which we do not current support extensions.
	return nil, errors.Join(ErrExtensionsUnsupportedForDBType, fmt.Errorf("database_type=%s", db.dbType))
}

func (db *BaseDatabase) checkForExtensions() {
	var msg string
	if len(db.extensions) == 0 {
		goto end
	}
	msg = "LoadExtension() not implemented for database type"
	db.Writer.Errorf("%s '%s'. Did you forget to implement?  Extensions:", msg, db.dbType)
	for _, ext := range db.extensions {
		db.Writer.Errorf("- %s", ext.Name())
	}
	db.Logger.Error(msg, "database_type", db.dbType, "extensions", db.extensions)
end:
}

func (db *BaseDatabase) Initialize(ctx Context) (err error) {
	if db.Initialized {
		goto end
	}
	err = db.parent.Open(ctx)
	db.Initialized = true
end:
	return err
}

func (db *BaseDatabase) ConnectString() string {
	return db.conn
}

func (db *BaseDatabase) Query(ctx Context, q string, params ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, q, params...)
}

// CheckFileConnection checks for file connections which work for SQLite3 and DuckDB.
func (db *BaseDatabase) CheckFileConnection(ctx Context, dbType DatabaseType, cs common.Filepath) (err error) {
	err = common.CheckFileExists(cs)
	switch {
	case errors.Is(err, common.ErrFileDoesNotExist):
		goto end
	case errors.Is(err, common.ErrPathIsDir):
		goto end
	}
	err = db.CheckDBConnection(ctx, dbType, common.ConnectString(cs))
end:
	return err
}

// CheckDBConnection checks for file connections which work for SQLite3 and DuckDB.
func (db *BaseDatabase) CheckDBConnection(_ Context, dbType DatabaseType, cs common.ConnectString) (err error) {
	var sqlDB *sql.DB

	defer func() {
		if err != nil {
			return
		}
		e := recover()
		if e != nil {
			err = e.(error)
		}
	}()
	sqlDB, err = sql.Open(string(dbType), string(cs))
	if err != nil {
		goto end
	}
	err = sqlDB.Ping()
	if err != nil {
		goto end
	}
	common.CloseOrLog(sqlDB) // DO NOT move this up and defer it; that will cascade errors
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
	return db.extensions
}
