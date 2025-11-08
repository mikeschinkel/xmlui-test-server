package dbpkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-doterr"
	"github.com/mikeschinkel/go-dt"
	"github.com/xmlui-org/localsvr/xmluisvr/cfgldr"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
	"github.com/xmlui-org/localsvr/xmluisvr/dbqvars"
	"github.com/xmlui-org/localsvr/xmluisvr/pathvars"

	. "github.com/mikeschinkel/go-doterr"
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

type FormatParamFunc = func(int) string

type Database interface {
	Type() DatabaseType
	TypeName() string
	ConnectString() string
	SetConnectString(string)
	SourceFile() dt.Filepath
	SetBaseDatabase(db *BaseDatabase)
	Open(Context) error
	Close() error
	Query(Context, string, ...any) (*sql.Rows, error)
	ValidatedConnection(Context, DatabaseType, common.ConnectString) error
	ParseConnectString(string) (common.ConnectString, error)
	ParseQueryString(query string) (dbqvars.QueryString, error)
	QueryFileExt() string
	ParseExtension(DBExtensionConfig) (DBExtension, error)
	CreateNew(DatabaseArgs) (Database, error)
	Extensions() []DBExtension
	LoadExtension(DBExtension) error
	GetFormatParamFunc() FormatParamFunc
	Options() common.Options
	ConvertValue(value any, dt dbqvars.DBDataType) any
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
	SourceFile       dt.Filepath
	CLIWriter        CLIWriter
	Logger           *slog.Logger
	Config           cfgldr.DatabaseConfig
}

type ParseQueriesArgs struct {
	Database     Database
	BaseFilename string
	ConfigSource dt.Filepath
	DirsProvider *cfgstore.DirsProvider
}

func ParseQueries(queries []string, args ParseQueriesArgs) (mpq *MultipartQuery, err error) {
	var queryBytes []byte
	var fileQuery string
	var elemCnt, lineCnt int
	var csFilepath dt.Filepath

	mpq = NewMultipartQuery()
	elemCnt = len(queries)
	for i, qs := range queries {
		mpq.AddQuerySource(
			NewQuerySource(i+1, i+1, common.QueryString(qs), args.ConfigSource),
		)
	}

	db := args.Database
	cs := cfgstore.NewConfigStore(cfgstore.CLIConfigDir, cfgstore.ConfigStoreArgs{
		ConfigSlug:   common.ConfigSlug,
		RelFilepath:  dt.RelFilepathJoin3(ConfigSlug, db.Type(), args.BaseFilename+db.QueryFileExt()),
		DirsProvider: args.DirsProvider,
	})
	//cs := cfgstore.NewConfigStore(common.AppConfigSlug,
	//	fmt.Sprintf("%s/%s%s", db.Type(), args.BaseFilename, db.QueryFileExt()),
	//	cfgstore.DefaultConfigDirType,
	//)
	queryBytes, err = cs.Load()
	if errors.Is(err, cfgstore.ErrFileDoesNotExist) {
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
			dt.Filepath(csFilepath),
		),
	)
end:
	return mpq, err
}

type ParseDatabaseArgs struct {
	Options      *common.Options
	Writer       CLIWriter
	Logger       *slog.Logger
	DirsProvider *cfgstore.DirsProvider
}

var ErrNoDatabaseConnectString = errors.New("no database connection string")

func ParseDatabase(ctx Context, cfg cfgldr.DatabaseConfig, args ParseDatabaseArgs) (db Database, err error) {
	var dbType DatabaseType
	var exts []DBExtension
	var bootstrapQueries, onOpenQueries *MultipartQuery
	var sourceFile dt.Filepath

	switch {
	case cfg.DatabaseType() != "":
		dbType = DatabaseType(cfg.DatabaseType())
	case args.Options.ConnectString != "":
		dbType, err = ParseDatabaseType(ctx, string(args.Options.ConnectString))
	case cfg.ConnectString() != "":
		dbType, err = ParseDatabaseType(ctx, cfg.ConnectString())
	default:
		err = NewErr(ErrNoDatabaseConnectString)
	}
	if err != nil {
		goto end
	}

	db, err = GetRegisteredDatabase(dbType)
	if db == nil {
		err = NewErr(ErrUnsupportedDBType, "database_type", dbType, err)
		goto end
	}

	exts, err = ParseExtensions(db, cfg.DBExtensions())
	if err != nil {
		goto end
	}

	sourceFile, err = dt.ParseFilepath(cfg.SourceFile())
	if err != nil {
		goto end
	}

	bootstrapQueries, err = ParseQueries(cfg.BootstrapQueries(), ParseQueriesArgs{
		Database:     db,
		BaseFilename: "bootstrap",
		ConfigSource: sourceFile,
		DirsProvider: args.DirsProvider,
	})
	if err != nil {
		goto end
	}

	onOpenQueries, err = ParseQueries(cfg.OnOpenQueries(), ParseQueriesArgs{
		Database:     db,
		BaseFilename: "on_open",
		ConfigSource: sourceFile,
		DirsProvider: args.DirsProvider,
	})
	if err != nil {
		goto end
	}

	db, err = db.CreateNew(DatabaseArgs{
		DatabaseType:     dbType,
		ConnectString:    cfg.ConnectString(),
		Port:             cfg.Port(),
		Extensions:       exts,
		BootstrapQueries: bootstrapQueries,
		OnOpenQueries:    onOpenQueries,
		AccessMode:       ReadOnlyMode, //TODO: Make this configurable somehow
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
			errs = append(errs, NewErr(
				pathvars.ErrParseFailed,
				pathvars.ErrParsingDBExtensionFailed,
				"database_type", db.Type(),
				"extension", ext,
				err,
			))
			continue
		}
		dbExts = append(dbExts, dbExt)
	}
	return dbExts, doterr.CombineErrs(errs)
}
