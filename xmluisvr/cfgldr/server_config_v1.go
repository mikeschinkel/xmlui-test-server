package cfgldr

const (
	ServerSchemaVersion = 1
	ServerSchema        = "https://schemas.xmlui.org/v1/test-server-server-schema.json"
)

// ServerConfigV1 represents server-level configuration
type ServerConfigV1 struct {
	Schema        string       `json:"$schema,omitempty"`
	SchemaVersion int          `json:"$schemaVersion,omitempty"`
	Host          string       `json:"host,omitempty"`
	Port          int          `json:"port,omitempty"`
	API           *APIConfigV2 `json:"api,omitempty"`
}
type ServerConfigV1Args struct {
	Port int
	API  *APIConfigV2
}

func NewServerConfigV1(host string, args ServerConfigV1Args) *ServerConfigV1 {
	return &ServerConfigV1{
		Schema:        ServerSchema,
		SchemaVersion: ServerSchemaVersion,
		Host:          host,
		Port:          args.Port,
		API:           args.API,
	}
}
