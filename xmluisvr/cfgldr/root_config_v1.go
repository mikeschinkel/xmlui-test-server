package cfgldr

import (
	"encoding/json"

	"github.com/xmlui-org/xmluisvr/cfgutil"
)

// TODO — Convince Gent that we should publish schemas on schemas.xmlui.org
const (
	RootSchemaVersion = 1
	RootConfigFile    = "test-server.json"
	RootSchema        = "https://schemas.xmlui.org/v1/test-server-root-schema.json"
)

var _ Config = (*RootConfigV1)(nil)

// RootConfigV1 represents the root configuration structure as defined in ADR-001
type RootConfigV1 struct {
	Schema        string          `json:"$schema"`
	SchemaVersion int             `json:"$schemaVersion"`
	Server        *ServerConfigV1 `json:"server"`
	Database      DatabaseConfig  `json:"database"`
}

func NewRootConfigV1() *RootConfigV1 {
	return &RootConfigV1{
		Schema:        RootSchema,
		SchemaVersion: RootSchemaVersion,
	}
}
func (c *RootConfigV1) Config() {}
func (c *RootConfigV1) String() string {
	return string(c.Bytes())
}
func (c *RootConfigV1) Bytes() []byte {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		panic(err)
	}
	return b
}

func (c *RootConfigV1) UnmarshalJSON(data []byte) (err error) {
	var dbc DatabaseConfig

	// Create a temporary struct that matches RootConfigV1 but with Database as RawMessage
	var temp struct {
		Schema        string          `json:"$schema"`
		SchemaVersion int             `json:"$schemaVersion"`
		Server        *ServerConfigV1 `json:"server"`
		Database      json.RawMessage `json:"database"`
	}

	var typeInfo struct {
		Type DatabaseType `json:"type"`
	}

	err = json.Unmarshal(data, &temp)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(temp.Database, &typeInfo)
	if err != nil {
		goto end
	}

	dbc, err = GetDatabaseConfig(typeInfo.Type)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(temp.Database, &dbc)
	if err != nil {
		goto end
	}

	c.Schema = temp.Schema
	c.SchemaVersion = temp.SchemaVersion
	c.Server = temp.Server
	c.Database = dbc

end:
	return err
}

func LoadRootConfigV1(appName string) (rc *RootConfigV1, err error) {
	return LoadRootConfigV1FromConfigStoreMap(
		cfgutil.GetConfigStoreDirTypeMap(appName, RootConfigFile),
	)
}

// LoadRootConfigV1FromConfigStoreMap also specifying the config stores in a map to enable unit testing
func LoadRootConfigV1FromConfigStoreMap(stores cfgutil.ConfigStoreDirTypeMap) (rc *RootConfigV1, err error) {
	var localConfig RootConfigV1
	var cs cfgutil.ConfigStore
	var fp string

	userConfig := RootConfigV1{}
	cs = stores[cfgutil.DotConfigDir]
	err = cs.LoadJSON(&userConfig)
	if err != nil {
		goto end
	}
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	userConfig.Server.API.SourceFile = fp

	localConfig = RootConfigV1{}
	cs = stores[cfgutil.LocalConfigDir]
	err = cs.LoadJSON(&localConfig)
	if err != nil {
		goto end
	}
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	localConfig.Server.API.SourceFile = fp

	// TODO Merge them here instead of just returning userConfig
	rc = &userConfig
end:
	return rc, err
}
