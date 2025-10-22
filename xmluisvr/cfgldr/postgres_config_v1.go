package cfgldr

var _ DatabaseConfig = (*PostgresConfigV1)(nil)
var _ Config = (*PostgresConfigV1)(nil)

type PostgresConfigV1 struct {
	Notes []string `json:"@notes,omitempty"`
	URL   string   `json:"pg_url"`
}

func (c *PostgresConfigV1) Clone() DatabaseConfig {
	newPG := *c
	return &newPG
}

func (c *PostgresConfigV1) SetBootstrapQueries(strings []string) {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) BootstrapQueries() []string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) OnOpenQueries() []string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) SourceFile() string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) DatabaseConfig() {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) DatabaseType() DatabaseType {
	return PostgresDatabase
}

func (c *PostgresConfigV1) SetConnectString(_ string) {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) ConnectString() string {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) Port() int {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) DBExtensions() []DBExtensionConfig {
	//TODO implement me
	panic("implement me")
}

func (c *PostgresConfigV1) Normalize(sourceFile string, opts *Options) {
	//TODO implement me
	panic("implement me")
}

func (*PostgresConfigV1) Config() {}
