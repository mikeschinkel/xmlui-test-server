package dbpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

//type ConnectStyle string
//const (
//	FileConnect ConnectStyle = "file"
//	URLConnect ConnectStyle = "url"
//	DSNConnect ConnectStyle = "dsn"
//	URLOrDSNConnect ConnectStyle = "url|dsn"
//)
//ConnectStyle() ConnectStyle

type DBExtensionConfig interface {
	DBExtensionConfig()
}

type Database interface {
	Type() DatabaseType
	TypeName() string
	ConnectString() string
	SourceFile() common.Filepath
	SetBaseDatabase(db *BaseDatabase)
	Open(Context) error
	Query(Context, string, ...any) (*sql.Rows, error)
	CheckConnection(Context, DatabaseType, common.ConnectString) error
	ParseConnectString(string) (common.ConnectString, error)
	ParseQueryString(query string) (common.QueryString, error)
	QueryFileExt() string
	ParseExtension(DBExtensionConfig) (DBExtension, error)
	CreateNew(DatabaseArgs) (Database, error)
	Extensions() []DBExtension
	LoadExtension(DBExtension) error
	fmt.Stringer
}

type DatabaseArgs struct {
	DatabaseType     DatabaseType
	ConnectString    string
	Port             int
	Extensions       []DBExtension
	BootstrapQueries *MultipartQuery
	OnOpenQueries    *MultipartQuery
	Options          *common.Options
	AccessMode       AccessMode
	SourceFile       common.Filepath
	CLIWriter        CLIWriter
	Logger           *slog.Logger
	Config           cfgldr.DatabaseConfig
}

type ParseQueriesArgs struct {
	Database     Database
	BaseFilename string
	ConfigSource common.Filepath
}

func ParseQueries(queries []string, args ParseQueriesArgs) (mpq *MultipartQuery, err error) {
	var queryBytes []byte
	var fileQuery string
	var elemCnt, lineCnt int
	var csFilepath string

	mpq = NewMultipartQuery()
	elemCnt = len(queries)
	for i, qs := range queries {
		mpq.AddQuerySource(
			NewQuerySource(i+1, i+1, common.QueryString(qs), args.ConfigSource),
		)
	}

	db := args.Database
	cs := cfgutil.NewConfigStoreWithFilename(common.AppConfigPath,
		fmt.Sprintf("%s/%s%s", db.Type(), args.BaseFilename, db.QueryFileExt()),
		cfgutil.DefaultConfigDirType,
	)
	queryBytes, err = cs.Load()
	if errors.Is(err, cfgutil.ErrFileDoesNotExist) {
		err = nil
	}
	if err != nil {
		goto end
	}
	fileQuery = strings.TrimSpace(string(queryBytes))
	if len(fileQuery) == 0 {
		goto end
	}
	lineCnt = strings.Count(fileQuery, "\n") + 1
	csFilepath, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	mpq.AddQuerySource(
		NewQuerySource(
			elemCnt+1,
			elemCnt+lineCnt,
			common.QueryString(fileQuery),
			common.Filepath(csFilepath),
		),
	)
end:
	return mpq, err
}

type ParseDatabaseArgs struct {
	Options *common.Options
	Writer  CLIWriter
	Logger  *slog.Logger
}

func ParseDatabase(ctx Context, cfg cfgldr.DatabaseConfig, args ParseDatabaseArgs) (db Database, err error) {
	var dt DatabaseType
	var exts []DBExtension
	var bootstrapQueries, onOpenQueries *MultipartQuery
	var sourceFile common.Filepath

	dt, err = ParseDatabaseType(ctx, cfg.ConnectString())
	if err != nil {
		goto end
	}

	db, err = GetRegisteredDatabase(dt)
	if db == nil {
		err = errors.Join(ErrUnsupportedDBType, fmt.Errorf("database_type=%s", dt), err)
		goto end
	}

	exts, err = ParseExtensions(db, cfg.DBExtensions())
	if err != nil {
		goto end
	}

	sourceFile, err = common.ParseFilepath(cfg.SourceFile())
	if err != nil {
		goto end
	}

	bootstrapQueries, err = ParseQueries(cfg.BootstrapQueries(), ParseQueriesArgs{
		Database:     db,
		BaseFilename: "bootstrap",
		ConfigSource: sourceFile,
	})
	if err != nil {
		goto end
	}

	onOpenQueries, err = ParseQueries(cfg.OnOpenQueries(), ParseQueriesArgs{
		Database:     db,
		BaseFilename: "on_open",
		ConfigSource: sourceFile,
	})
	if err != nil {
		goto end
	}

	db, err = db.CreateNew(DatabaseArgs{
		DatabaseType:     dt,
		ConnectString:    cfg.ConnectString(),
		Port:             cfg.Port(),
		Extensions:       exts,
		BootstrapQueries: bootstrapQueries,
		OnOpenQueries:    onOpenQueries,
		AccessMode:       0,
		SourceFile:       sourceFile,
		Config:           cfg,
		Options:          args.Options,
		CLIWriter:        args.Writer,
		Logger:           args.Logger,
	})
end:
	return db, err
}

type DBExtension interface {
	Name() string
	DBExtension()
}

func ParseExtensions(db Database, exts []cfgldr.DBExtensionConfig) (dbExts []DBExtension, err error) {
	var errs []error
	var dbExt DBExtension
	dbExts = make([]DBExtension, 0, len(exts))
	for _, ext := range exts {
		dbExt, err = db.ParseExtension(ext)
		if err != nil {
			errs = append(errs, err,
				fmt.Errorf("database_type=%s", db.Type()),
				fmt.Errorf("extension=%s", ext),
			)
			continue
		}
		dbExts = append(dbExts, dbExt)
	}
	return dbExts, errors.Join(errs...)
}
