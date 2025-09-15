package cfgldr

import (
	"errors"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

const (
	ServerConfigV1SchemaVersion = 1
	ServerConfigV1Schema        = "https://schemas.xmlui.org/v1/test-server-server-schema.json"
)

type ServerConfig interface {
	ServerConfig()
}

var _ ServerConfig = (*ServerConfigV1)(nil)

// ServerConfigV1 represents server-level configuration
type ServerConfigV1 struct {
	Schema        string       `json:"$schema,omitempty"`
	SchemaVersion int          `json:"$schemaVersion,omitempty"`
	Host          string       `json:"host,omitempty"`
	Port          int          `json:"port,omitempty"`
	APIConfig     *APIConfigV2 `json:"api,omitempty"`
	SourceFile    string       `json:"-"`
}

func (*ServerConfigV1) ServerConfig() {}

func (c *ServerConfigV1) Normalize(sourceFile string) (err error) {
	var errs []error
	c.Schema = ServerConfigV1Schema
	c.SchemaVersion = ServerConfigV1SchemaVersion
	c.SourceFile = sourceFile
	host, err := common.ParseHost(c.Host)
	if err != nil {
		errs = append(errs, err)
	}
	if host != "" {
		c.Host = string(host)
	}
	var port common.ServerPort
	port, err = common.ParseServerPort(c.Port, common.ZeroInvalid)
	if err != nil {
		errs = append(errs, err)
	}
	c.Port = int(port)
	err = c.APIConfig.Normalize(sourceFile)
	return errors.Join(errs...)
}

type ServerConfigV1Args struct {
	Port int
	API  *APIConfigV2
}

func NewServerConfigV1(host string, args ServerConfigV1Args) *ServerConfigV1 {
	return &ServerConfigV1{
		Schema:        ServerConfigV1Schema,
		SchemaVersion: ServerConfigV1SchemaVersion,
		Host:          host,
		Port:          args.Port,
		APIConfig:     args.API,
	}
}
