package cfgldr

const UnknownVersion = "v0.0.0"

const (
	DefaultSQLite3ExtensionEntryPoint = "sqlite3_extension_init"
	DefaultOnFailurePolicy            = "warn"
	DefaultVarScope                   = "app"
	DefaultSQLite3Database            = "./data.db"
	DefaultAPIBasePath                = "/api"
	DefaultAPIWebroot                 = "."
)
