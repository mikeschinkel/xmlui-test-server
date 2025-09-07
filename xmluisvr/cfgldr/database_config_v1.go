package cfgldr

type DatabaseConfig interface {
	DatabaseConfig() // Marker
	DatabaseType() DatabaseType
	ConnectString() string
	Port() int
	Extensions() []string // TODO Change this to load all extension info

}
