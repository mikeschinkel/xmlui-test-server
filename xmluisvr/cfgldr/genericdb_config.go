package cfgldr

func init() {
	registerDatabaseConfig(&GenericDBConfig{})
}

var _ DatabaseConfig = (*GenericDBConfig)(nil)

type GenericDBConfig struct {
	connectString string
	port          int
	extensions    []string
}

func NewGenericDBConfig(args GenericDBConfigArgs) *GenericDBConfig {
	return &GenericDBConfig{
		connectString: args.ConnectString,
		port:          args.Port,
		extensions:    args.Extensions}
}

type GenericDBConfigArgs struct {
	ConnectString string
	Port          int
	Extensions    []string
}

func (g GenericDBConfig) DatabaseConfig() {
}

func (g GenericDBConfig) DatabaseType() DatabaseType {
	return GenericDatabase
}

func (g GenericDBConfig) ConnectString() string {
	return g.connectString
}

func (g GenericDBConfig) Port() int {
	return g.port
}

func (g GenericDBConfig) Extensions() []string {
	return g.extensions
}
