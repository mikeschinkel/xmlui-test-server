package cfgldr

var _ Config = (*PostgresConfig)(nil)

type PostgresConfig struct {
	URL string `json:"pg_url"`
}

func (*PostgresConfig) Config() {}
