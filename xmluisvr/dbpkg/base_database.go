package dbpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type BaseDatabase struct {
	*sql.DB
	cliutil.WriterLogger
	dbType           DatabaseType
	conn             string
	parent           Database
	BootstrapQueries *MultipartQuery
	OnOpenQueries    *MultipartQuery
	extensions       []DBExtension
	sourceFile       common.Filepath
	options          *common.Options
	AccessMode       AccessMode
	Initialized      bool
}

func (db *BaseDatabase) Options() common.Options {
	return *db.options
}

func (db *BaseDatabase) String() string {
	//TODO implement me
	panic("implement me")
}

func NewBaseDatabase(parent Database, args DatabaseArgs) *BaseDatabase {
	if args.AccessMode == UnspecifiedAccessMode {
		args.AccessMode = ReadWriteMode
	}
	return &BaseDatabase{
		parent:           parent,
		dbType:           args.DatabaseType,
		conn:             args.ConnectString,
		BootstrapQueries: args.BootstrapQueries,
		OnOpenQueries:    args.OnOpenQueries,
		extensions:       args.Extensions,
		options:          args.Options,
		AccessMode:       args.AccessMode,
		sourceFile:       args.SourceFile,
		WriterLogger:     cliutil.NewWriterLogger(args.CLIWriter, args.Logger),
	}
}

func (db *BaseDatabase) Close() error {
	return db.DB.Close()
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

func (db *BaseDatabase) SetConnectString(cs string) {
	db.conn = cs
}

func (db *BaseDatabase) Query(ctx Context, q string, params ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, q, params...)
}

type MissingFileOpenMode int

const (
	UnspecifiedCreateMode MissingFileOpenMode = iota
	CreatesMissingFileOnOpen
	FailsOnOpenOfMissingFile
)

// CheckFileConnection checks for file connections which work for SQLite3 and DuckDB.
func (db *BaseDatabase) CheckFileConnection(ctx Context, dbType DatabaseType, cs common.Filepath, mode MissingFileOpenMode) (err error) {
	err = common.CheckFileExists(cs)
	switch {
	case errors.Is(err, common.ErrFileDoesNotExist):
		if mode != CreatesMissingFileOnOpen {
			goto end
		}
		// Calls says it will be created on open
		err = common.EnsureDirExists(common.Dir(cs))
		if err != nil {
			goto end
		}
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
	return common.HomeRelative(absPath)
}

func (db *BaseDatabase) Extensions() []DBExtension {
	return db.extensions
}
