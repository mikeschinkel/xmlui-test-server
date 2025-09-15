package cfgldr

var _ DatabaseConfig = (*PostgresConfig)(nil)
var _ Config = (*PostgresConfig)(nil)

type PostgresConfig struct {
	URL string `json:"pg_url"`
}

func (c *PostgresConfig) SetSchemaQueries(strings []string) {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) SchemaQueries() []string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) OnOpenQueries() []string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) SourceFile() string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) DatabaseConfig() {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) DatabaseType() DatabaseType {
	return PostgresDatabase
}

func (c *PostgresConfig) ConnectString() string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) Port() int {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) DBExtensions() []DBExtensionConfig {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) Normalize(sourceFile string) error {
	//TODO implement me
	panic("implement me")
}

func (*PostgresConfig) Config() {}
