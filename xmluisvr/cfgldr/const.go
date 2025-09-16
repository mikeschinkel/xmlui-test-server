package cfgldr

const UnknownVersion = "v0.0.0"

const (
	DefaultSQLite3ExtensionEntryPoint = "sqlite3_extension_init"
	DefaultOnFailurePolicy            = "warn"
	DefaultVarScope                   = "app"
	DefaultAPIBasePath                = "/api"
	DefaultAPIWebroot                 = "./webroot"
	DefaultSQLite3Database            = "./data/data.db"
)
