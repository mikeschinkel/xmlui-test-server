package cfgldr

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

const (
	ServerConfigV1Version = 1
	ServerConfigV1Schema  = "https://schemas.xmlui.org/v1/test-server/server-schema.json"
)

type ServerConfig interface {
	ServerConfig()
}

var _ ServerConfig = (*ServerConfigV1)(nil)

// ServerConfigV1 represents server-level configuration
type ServerConfigV1 struct {
	Schema     string       `json:"$schema,omitempty"`
	Version    int          `json:"version,omitempty"`
	Host       string       `json:"host,omitempty"`
	Port       int          `json:"port,omitempty"`
	APIConfig  *APIConfigV2 `json:"api,omitempty"`
	SourceFile string       `json:"-"`
}

func (*ServerConfigV1) ServerConfig() {}

func (c *ServerConfigV1) Normalize(sourceFile string) {
	c.Schema = ServerConfigV1Schema
	c.Version = ServerConfigV1Version
	c.SourceFile = sourceFile
	if c.Host != "" {
		c.Host = common.DefaultServerHost
	}
	if c.Port == 0 {
		c.Port = common.DefaultServerPort
	}
	c.APIConfig.Normalize(sourceFile)
	return
}

type ServerConfigV1Args struct {
	Port int
	API  *APIConfigV2
}

func NewServerConfigV1(host string, args ServerConfigV1Args) *ServerConfigV1 {
	return &ServerConfigV1{
		Schema:    ServerConfigV1Schema,
		Version:   ServerConfigV1Version,
		Host:      host,
		Port:      args.Port,
		APIConfig: args.API,
	}
}
