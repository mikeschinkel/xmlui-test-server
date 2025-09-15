package cfgldr

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// TODO — Convince Gent that we should publish schemas on schemas.xmlui.org
const (
	RootConfigV1SchemaVersion = 1
	RootConfigFile            = "test-server.json"
	RootConfigV1Schema        = "https://schemas.xmlui.org/v1/test-server-root-schema.json"
)

var _ Config = (*RootConfigV1)(nil)

// RootConfigV1 represents the root configuration structure as defined in ADR-001
type RootConfigV1 struct {
	Schema        string          `json:"$schema"`
	SchemaVersion int             `json:"$schemaVersion"`
	ServerConfig  *ServerConfigV1 `json:"server"`
	DBConfig      DatabaseConfig  `json:"database"`
}

type RootConfigV1Args struct {
	ServerConfig *ServerConfigV1
	DBConfig     DatabaseConfig
}

func NewRootConfigV1(args RootConfigV1Args) *RootConfigV1 {
	return &RootConfigV1{
		Schema:        RootConfigV1Schema,
		SchemaVersion: RootConfigV1SchemaVersion,
		ServerConfig:  args.ServerConfig,
		DBConfig:      args.DBConfig,
	}
}

func (c *RootConfigV1) APIConfig() (ac APIConfig) {
	if c == nil {
		goto end
	}
	if c.ServerConfig == nil {
		goto end
	}
	ac = c.ServerConfig.APIConfig
end:
	return ac
}
func (c *RootConfigV1) Config() {}
func (c *RootConfigV1) Normalize(sourceFile string) (err error) {
	var errs []error
	c.Schema = RootConfigV1Schema
	c.SchemaVersion = RootConfigV1SchemaVersion
	err = c.ServerConfig.Normalize(sourceFile)
	if err != nil {
		errs = append(errs, err)
	}
	err = c.DBConfig.Normalize(sourceFile)
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

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

	// Create a temporary struct that matches RootConfigV1 but with DBConfig as RawMessage
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
	c.ServerConfig = temp.Server
	c.DBConfig = dbc

end:
	return err
}

func LoadRootConfigV1(appName string) (rc *RootConfigV1, err error) {
	typeMap := cfgutil.GetConfigStoreDirTypeMap(appName, RootConfigFile)
	return LoadRootConfigV1FromConfigStoreMap(typeMap)
}
func ensureConfig(cs cfgutil.ConfigStore) (rc *RootConfigV1, err error) {
	rc, err = maybeLoadConfig(cs)
	if err != nil {
		// A real error occurred, bail out
		goto end
	}
	if rc != nil {
		// Config was loaded, no need to create config
		goto end
	}
	rc, err = createConfig(cs)
end:
	return rc, err
}

func createConfig(cs cfgutil.ConfigStore) (rc *RootConfigV1, err error) {
	var api *APIConfigV2
	var db *SQLite3ConfigV1
	var server *ServerConfigV1
	var fp string

	api = NewAPIConfigV2(DefaultAPIWebroot)
	api.AddEndpoint(NewAPIEndpointV2("GET /hello", APIEndpointV2Args{
		Description: "Hello World Endpoint",
		Query:       "SELECT 'Hello World';",
		Cardinality: string(common.OneRow),
		RowType:     string(common.StringRowType),
	}))
	db = NewSQLite3ConfigV1("data.db")
	//err = db.AddExtension(common.AppConfigPath, &SQLite3ExtensionConfigV1{
	//	Filepath: "steampipe_sqlite_github.so",
	//})
	//if err != nil {
	//	goto end
	//}
	server = NewServerConfigV1(common.LocalHostIP, ServerConfigV1Args{
		Port: 8080,
		API:  api,
	})
	rc = NewRootConfigV1(RootConfigV1Args{
		ServerConfig: server,
		DBConfig:     db,
	})
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.Normalize(fp)
	if err != nil {
		goto end
	}
	err = cs.SaveJSON(rc)
	if err != nil {
		goto end
	}
end:
	return rc, err
}

func maybeLoadConfig(cs cfgutil.ConfigStore) (rc *RootConfigV1, err error) {
	var fp string
	if !cs.Exists() {
		goto end
	}
	rc = &RootConfigV1{}
	err = cs.LoadJSON(&rc)
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.Normalize(fp)
end:
	return rc, err
}

var (
	ErrFailedToLoadAPIConfigFile      = errors.New("failed to load APIConfig config file")
	ErrFailedToUnmarshalAPIConfigFile = errors.New("failed to unmarshal APIConfig config file")
	ErrFailedToLoadDBSchemaFile       = errors.New("failed to load DB schema file")
)

// LoadRootConfigV1FromConfigStoreMap also specifying the config stores in a map to enable unit testing
func LoadRootConfigV1FromConfigStoreMap(stores cfgutil.ConfigStoreDirTypeMap) (rc *RootConfigV1, err error) {
	var userConfig, localConfig *RootConfigV1
	var cs cfgutil.ConfigStore
	var schemaBytes []byte
	var apiConfig *APIConfigV2
	var opts *Options

	cs = stores[cfgutil.DotConfigDir]
	userConfig, err = ensureConfig(cs)
	if err != nil {
		goto end
	}

	cs = stores[cfgutil.LocalConfigDir]
	localConfig, err = maybeLoadConfig(cs)
	if err != nil {
		goto end
	}

	// TODO Merge them here instead of just returning userConfig
	common.Noop(localConfig)
	rc = userConfig

	opts = GetOptions()
	apiConfig, err = loadAPIFileIfExists(opts.APIFile)
	if err != nil {
		goto end
	}
	if apiConfig != nil {
		rc.ServerConfig.APIConfig = apiConfig
	}
	if rc.DBConfig == nil {
		rc.DBConfig = NewSQLite3ConfigV1(DefaultSQLite3Database)
	}

	schemaBytes, err = cfgutil.ReadFileIfExists(opts.DBSchemaFile)
	if err != nil {
		err = errors.Join(ErrFailedToLoadDBSchemaFile, fmt.Errorf("dbschema_file=%s", opts.DBSchemaFile), err)
		goto end
	}
	if len(schemaBytes) != 0 {
		rc.DBConfig.SetSchemaQueries([]string{string(schemaBytes)})
	}

end:
	return rc, err
}

func loadAPIFileIfExists(apiFile string) (api *APIConfigV2, err error) {
	var errs [2]error
	var apiBytes []byte
	if apiFile == "" {
		goto end
	}
	apiBytes, err = cfgutil.ReadFileIfExists(apiFile)
	if err != nil {
		errs = [2]error{ErrFailedToLoadAPIConfigFile, err}
		goto end
	}
	api = &APIConfigV2{}
	err = json.Unmarshal(apiBytes, &api)
	if err != nil {
		errs = [2]error{ErrFailedToUnmarshalAPIConfigFile, err}
		goto end
	}
end:
	err = nil
	if errs[0] != nil {
		err = errors.Join(errs[0], fmt.Errorf("api_file=%s", apiFile), errs[1])
	}
	return api, err
}
