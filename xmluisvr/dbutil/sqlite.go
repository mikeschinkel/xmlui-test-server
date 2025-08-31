package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/xmlui-org/xmluisvr/fsutil"
)

var _ Database = (*SQLiteDB)(nil)

type SQLiteDB struct {
	ExtensionPaths []string
	*database
}

func (SQLiteDB) Type() DatabaseType {
	return SQLiteDatabase
}

func NewSqliteDB(args DatabaseArgs) *SQLiteDB {
	return &SQLiteDB{
		database:       newDatabase(args),
		ExtensionPaths: args.ExtensionPaths,
	}
}

func (s SQLiteDB) Name() string {
	return s.homeRelativeFile()
}

func (s SQLiteDB) homeRelativeFile() string {
	absPath, err := filepath.Abs(s.conn)
	if err != nil {
		panic(fmt.Sprintf("Failed to get absolute path of '%s': %v", s.conn, err))
	}
	return fsutil.HomeRelative(absPath)
}

func (s SQLiteDB) Open() (err error) {
	// Default to SQLite
	s.logger.Info("Opening SQLite DB: %s\n", s.homeRelativeFile())
	// Simple connection string with extension loading enabled
	s.db, err = sql.Open("sqlite3", s.conn+"?_allow_load_extension=1")
	if err != nil {
		err = errors.Join(ErrConnFailed, err)
		goto end
	}

	// SQLite specific configurations
	s.db.SetMaxOpenConns(1)
	s.db.SetMaxIdleConns(1)

	// Create memory database for extensions
	_, err = s.db.Exec(`ATTACH DATABASE ':memory:' AS extension_mem`)
	if err != nil {
		// TODO: Do we want to fail to run the server or allow without extension?
		s.writer.Errorf("Failed to attach memory database for extension loading")
		s.logger.Warn("Failed to attach memory database", "error", err)
		goto end
	}

	// Enable extension loading via PRAGMA
	_, err = s.db.Exec(`PRAGMA load_extension = 1;`)
	if err != nil {
		// TODO: Do we want to fail to run the server or allow without extension?
		s.writer.Errorf("Failed to enable extension loading")
		s.logger.Warn("PRAGMA load_extension failed", "error", err)
	}
	if !s.hasExtensions() {
		goto end
	}
	// If extension is provided, try to load it
	err = s.loadExtensions()
	if err != nil {
		goto end
	}

	err = os.Setenv("STEAMPIPE_CACHE", "false")
	if err != nil {
		s.writer.Errorf("Failed to set STEAMPIPE_CACHE: %s", err)
		s.logger.Error("Failed to set STEAMPIPE_CACHE", "error", err)
	}

end:
	return err
}

func (s SQLiteDB) hasExtensions() bool {
	return len(s.ExtensionPaths) > 0
}

func (s SQLiteDB) loadExtensions() error {
	var errs = make([]error, 0)
	for _, path := range s.ExtensionPaths {
		errs = append(errs, s.loadExtension(path))
	}
	return errors.Join(errs...)
}

func (s SQLiteDB) loadExtension(path string) (err error) {
	var absPath string
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	// Get the absolute path to the extension file
	absPath, err = filepath.Abs(path)
	if err != nil {
		s.logger.Warn("Failed to get absolute path for extension", "error", err)
		absPath = filepath.Join("./", path)
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
	_, err = s.db.Exec(loadQuery)
	if err != nil {
		s.logger.Warn("Extension loading failed with", "error", err)
		s.writer.Errorf("Extension failed to loaded\n")
		goto end
	}
	s.writer.Printf("Extension loaded successfully\n")

end:
	return err
}
