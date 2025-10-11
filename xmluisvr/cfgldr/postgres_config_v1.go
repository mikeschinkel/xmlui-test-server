package cfgldr

var _ DatabaseConfig = (*PostgresConfig)(nil)
var _ Config = (*PostgresConfig)(nil)

type PostgresConfig struct {
	URL string `json:"pg_url"`
}

func (c *PostgresConfig) Clone() DatabaseConfig {
	newPG := *c
	return &newPG
}

func (c *PostgresConfig) SetBootstrapQueries(strings []string) {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfig) BootstrapQueries() []string {
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

func (c *PostgresConfig) SetConnectString(_ string) {
	//TODO implement me
	panic("implement me")
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

func (c *PostgresConfig) Normalize(sourceFile string, opts *Options) {
	//TODO implement me
	panic("implement me")
}

func (*PostgresConfig) Config() {}
