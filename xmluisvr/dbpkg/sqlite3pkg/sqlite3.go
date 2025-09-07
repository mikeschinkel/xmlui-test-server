package sqlite3pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mattn/go-sqlite3"
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

const SQLite3Database dbpkg.DatabaseType = "sqlite3"

func init() {
	dbpkg.RegisterDatabase(&SQLite3{})
}

var _ dbpkg.Database = (*SQLite3)(nil)

type database = dbpkg.BaseDatabase

type SQLite3 struct {
	*database
	extensions  []*Extension
	initSQL     []string
	initialized bool
	logger      *slog.Logger
	writer      cliutil.Writer
}

func NewSqlite3(args dbpkg.DatabaseArgs) *SQLite3 {
	exts := make([]*Extension, 0, len(args.Extensions))
	for _, ext := range args.Extensions {
		exts = append(exts, ext.(*Extension))
	}
	db := &SQLite3{
		extensions: exts,
		logger:     args.Logger,
		writer:     args.CLIWriter,
	}
	db.database = dbpkg.NewBaseDatabase(db, args)
	return db
}

func (*SQLite3) CreateNew(args dbpkg.DatabaseArgs) dbpkg.Database {
	return NewSqlite3(args)
}

func (*SQLite3) Type() dbpkg.DatabaseType {
	return SQLite3Database
}

func (s *SQLite3) TypeName() string {
	return "SQLite3"
}

func (s *SQLite3) ParseExtensions(files []string) (exts []dbpkg.DBExtension, err error) {
	var errs []error
	exts = make([]dbpkg.DBExtension, 0, len(files))
	for _, file := range files {
		fileErr := fmt.Errorf("extension_file=%s", file)
		err = common.CheckFileExists(file)
		switch {
		case errors.Is(err, common.ErrFileDoesNotExist):
			err = errors.Join(fileErr, err)
		case errors.Is(err, common.ErrPathIsDir):
			err = errors.Join(fileErr, err)
		default:
			ext := NewExtension(common.Filepath(filepath.Base(file)))
			exts = append(exts, ext)
		}
		errs = append(errs, err)
	}
	return exts, errors.Join(errs...)
}

func (s *SQLite3) Extensions() (exts []dbpkg.DBExtension) {
	exts = make([]dbpkg.DBExtension, 0, len(s.extensions))
	for _, ext := range s.extensions {
		exts = append(exts, ext)
	}
	return exts
}

func (s *SQLite3) CheckConnection(cs string) (err error) {
	return s.CheckFileConnection(s, cs)
}

func (s *SQLite3) String() string {
	return s.HomeRelativeFile()
}

func (s *SQLite3) Open() (err error) {
	// Default to SQLite
	s.logger.Info("Opening SQLite DB", "db_file", s.HomeRelativeFile())
	if !s.initialized {
		sql.Register("sqlite3_ext",
			&sqlite3.SQLiteDriver{
				ConnectHook: s.ConnectHook(ExtensionOptions{
					AllowPragmaOverload: false, // stays off
				},
				),
			})
	}
	// Simple connection string with extension loading enabled
	s.DB, err = sql.Open("sqlite3", s.ConnectString()+"?_allow_load_extension=1")
	if err != nil {
		err = errors.Join(dbpkg.ErrConnFailed, err)
		goto end
	}

	// SQLite specific configurations
	s.DB.SetMaxOpenConns(1)
	s.DB.SetMaxIdleConns(1)

	// Create memory database for extensions
	// Concatenation used to stop IDE from flagging this an an error
	_, err = s.DB.Exec(`ATTACH ` + `DATABASE ':memory:' AS extension_mem`)
	if err != nil {
		// TODO: Do we want to fail to run the server or allow without extension?
		s.writer.Errorf("Failed to attach memory database for extension loading")
		s.logger.Warn("Failed to attach memory database", "error", err)
		goto end
	}

	panic("Handled these PRAGMAs somewhere")
	//PRAGMA journal_mode=WAL;
	//PRAGMA busy_timeout=5000;         -- ms; tune to your workload
	//PRAGMA synchronous=NORMAL;        -- or FULL if you want max durability

	if !s.hasExtensions() {
		goto end
	}
	// If extension is provided, try to load it
	err = s.loadExtensions()
	if err != nil {
		goto end
	}

end:
	return err
}

func (s *SQLite3) hasExtensions() bool {
	return len(s.extensions) > 0
}

func (s *SQLite3) loadExtensions() error {
	var errs = make([]error, 0)

	for _, ext := range s.extensions {
		errs = append(errs, s.loadExtension(ext))
	}
	return errors.Join(errs...)
}

func (s *SQLite3) ConnectHook(opts ExtensionOptions) func(*sqlite3.SQLiteConn) error {
	return func(conn *sqlite3.SQLiteConn) error {
		return s.Load(conn, opts)
	}
}

func (s *SQLite3) Load(conn *sqlite3.SQLiteConn, opts ExtensionOptions) (err error) {
	var errs []error
	queries := append([]string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}, s.initSQL...)
	for _, p := range queries {
		_, err := conn.Exec(p, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, ext := range s.extensions {
		errs = append(errs, ext.Load(conn, opts))
	}
	return errors.Join(errs...)
}

func (s *SQLite3) loadExtension(ext *Extension) (err error) {
	var absPath string
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	// TODO Check Filepath for URL and download if applicable

	filePath := string(ext.filePath)
	// Get the absolute path to the extension file
	absPath, err = filepath.Abs(filePath)
	if err != nil {
		s.logger.Warn("Failed to get absolute path for extension", "error", err)
		absPath = filepath.Join("./", filePath)
	}

	// Ensure file has execute permissions (required for Linux)
	err = os.Chmod(absPath, 0755)
	if err != nil {
		s.logger.Warn("Failed to set execute permissions on extension", "error", err)
	}

	// Log extension loading attempt
	s.writer.Printf("Trying to load extension: %s\n", absPath)
	s.logger.Info("Loading extension", "extension", absPath)

	// TODO: This is vulnerable to SQL injection; we should harden it
	loadQuery := fmt.Sprintf("SELECT load_extension('%s')", strings.ReplaceAll(absPath, "'", "''"))
	_, err = s.DB.Exec(loadQuery)
	if err != nil {
		s.logger.Warn("Extension loading failed with", "error", err)
		s.writer.Errorf("Extension failed to loaded\n")
		goto end
	}
	s.writer.Printf("Extension loaded successfully\n")

end:
	return err
}

func RegisterDriverWithHook(cfg HookCfg) {

	sql.Register("sqlite3_ext", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			// Pragmas per connection
			if _, err := conn.Exec("PRAGMA journal_mode=WAL", nil); err != nil {
				return err
			}
			if _, err := conn.Exec("PRAGMA busy_timeout=5000", nil); err != nil {
				return err
			}
			if _, err := conn.Exec("PRAGMA synchronous=NORMAL", nil); err != nil {
				return err
			}
			if _, err := conn.Exec("PRAGMA foreign_keys=ON", nil); err != nil {
				return err
			}

			// Optional: give every connection its own scratch schema
			if _, err := conn.Exec(`ATTACH DATABASE ':memory:' AS extension_mem`, nil); err != nil {
				return err
			}

			// LoadJSON your extensions (resolved paths)
			for _, ext := range cfg.Extensions {
				if err := conn.LoadExtension(ext.Path, ""); err != nil {
					return err
				}
				for _, q := range ext.OnLoadSQL {
					if _, err := conn.Exec(q, nil); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})

	sql.Register("sqlite3_ext", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			// (1) PRAGMAs per connection
			if cfg.UseWAL {
				if _, err := conn.Exec("PRAGMA journal_mode=WAL", nil); err != nil {
					return err
				}
			}
			if cfg.BusyTimeoutMS > 0 {
				if _, err := conn.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", cfg.BusyTimeoutMS), nil); err != nil {
					return err
				}
			}
			if cfg.Synchronous != "" {
				if _, err := conn.Exec("PRAGMA synchronous="+cfg.Synchronous, nil); err != nil {
					return err
				}
			}
			if _, err := conn.Exec("PRAGMA foreign_keys=ON", nil); err != nil {
				return err
			}

			// (3) LoadJSON extensions on this connection
			for _, ext := range cfg.Extensions {
				if err := conn.LoadExtension(ext.Path, ""); err != nil {
					return fmt.Errorf("load %s: %w", ext.Name, err)
				}
				for _, q := range ext.OnLoadSQL {
					if _, err := conn.Exec(q, nil); err != nil {
						return fmt.Errorf("%s onload: %w", ext.Name, err)
					}
				}
			}
			return nil
		},
	})
}
