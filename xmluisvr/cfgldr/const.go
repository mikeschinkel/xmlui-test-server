package cfgldr

import (
	"path/filepath"
)

const UnknownVersion = "v0.0.0"

const (
	DefaultSQLite3ExtensionEntryPoint = "sqlite3_extension_init"
	DefaultOnFailurePolicy            = "warn"
	DefaultVarScope                   = "app"
	DefaultAPIBasePath                = "/api"
	DefaultAPIConfigFile              = "./api.json"
	DefaultWebroot                    = "webroot"
	DefaultDBRoot                     = "dbroot"
	DefaultSQLite3DBFile              = "data.db"
	DefaultDBBootstrapFile            = "bootstrap.sql"
)

var (
	DefaultSQLite3Database     = filepath.Join(DefaultDBRoot, DefaultSQLite3DBFile)
	DefaultDBBootstrapFilepath = filepath.Join(DefaultDBRoot, DefaultDBBootstrapFile)
)
