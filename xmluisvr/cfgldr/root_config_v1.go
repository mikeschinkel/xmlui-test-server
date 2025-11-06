package cfgldr

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"os"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"

	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/xmlui-org/localdev/xmluisvr/common"
	"github.com/xmlui-org/localdev/xmluisvr/dbqvars"

	. "github.com/mikeschinkel/go-doterr"
)

// TODO — Convince Gent that we should publish schemas on schemas.xmlui.org
const (
	RootConfigV1Version = 1
	RootConfigFile      = common.ConfigFile
	RootConfigV1Schema  = "https://schemas.xmlui.org/v1/test-server/root-schema.json"
)

var _ Config = (*RootConfigV1)(nil)

// RootConfigV1 represents the root configuration structure as defined in ADR-001
type RootConfigV1 struct {
	rootConfigV1Base `json:",inline"`
	DBConfig         DatabaseConfig `json:"database"`
}

func (c *RootConfigV1) RootConfig() {}
func (c *RootConfigV1) IsNil() (isNil bool) {
	isNil = true
	if c == nil {
		goto end
	}
	if c.DBConfig == nil {
		goto end
	}
	if c.ServerConfig == nil {
		goto end
	}
	isNil = false
end:
	return isNil
}

// Base struct with non-polymorphic fields
type rootConfigV1Base struct {
	Schema       string          `json:"$schema"`
	Version      int             `json:"version"`
	ServerConfig *ServerConfigV1 `json:"server"`
}

type RootConfigV1Args struct {
	ServerConfig *ServerConfigV1
	DBConfig     DatabaseConfig
}

func NewRootConfigV1(args RootConfigV1Args) *RootConfigV1 {
	return &RootConfigV1{
		rootConfigV1Base: rootConfigV1Base{
			Schema:       RootConfigV1Schema,
			Version:      RootConfigV1Version,
			ServerConfig: args.ServerConfig,
		},
		DBConfig: args.DBConfig,
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

func (c *RootConfigV1) Normalize(sourceFile dt.Filepath, opts *Options) error {
	var errs []error
	c.Schema = RootConfigV1Schema
	c.Version = RootConfigV1Version
	if c.ServerConfig == nil {
		c.ServerConfig = NewServerConfigV1(common.DefaultServerHost, ServerConfigV1Args{
			Port: common.DefaultServerPort,
			API:  NewAPIConfigV2("."),
		})
	}
	errs = AppendErr(errs, c.ServerConfig.Normalize(sourceFile, opts))
	if c.DBConfig == nil {
		c.DBConfig = NewSQLite3ConfigV1(DefaultSQLite3Database)
	}
	errs = AppendErr(errs, c.DBConfig.Normalize(sourceFile, opts))
	return CombineErrs(errs)
}

func (c *RootConfigV1) String() string {
	return string(c.Bytes())
}

func (c *RootConfigV1) Bytes() []byte {
	b, err := jsonv2.Marshal(c, jsontext.WithIndent("  "))
	if err != nil {
		panic(err)
	}
	return b
}

func (c *RootConfigV1) UnmarshalJSON(data []byte) (err error) {
	var dbc DatabaseConfig

	// Create temp struct with inline base and RawMessage for polymorphic field
	var temp struct {
		rootConfigV1Base `json:",inline"`
		Database         jsontext.Value `json:"database"`
	}

	var typeInfo struct {
		Type DatabaseType `json:"type"`
	}

	err = jsonv2.Unmarshal(data, &temp)
	if err != nil {
		goto end
	}

	// Copy non-polymorphic fields from temp
	c.rootConfigV1Base = temp.rootConfigV1Base

	// Handle polymorphic database field
	err = jsonv2.Unmarshal(temp.Database, &typeInfo)
	if err != nil {
		goto end
	}

	dbc, err = GetDatabaseConfig(typeInfo.Type)
	if err != nil {
		goto end
	}

	dbc = dbc.Clone()

	err = jsonv2.Unmarshal(temp.Database, &dbc)
	if err != nil {
		goto end
	}

	c.DBConfig = dbc

end:
	return err
}

func readFile(file string, mustLoad bool) (data []byte, err error) {
	data, err = os.ReadFile(file)
	if err != nil && !mustLoad {
		err = nil
	}
	return data, err
}

type LoadRootConfigV1Args struct {
	AppInfo      appinfo.AppInfo
	Options      cfgstore.Options
	DirTypes     []cfgstore.DirType
	ConfigStores *cfgstore.ConfigStores
}

var _ cfgstore.RootConfig = (*RootConfigV1Wrapper)(nil)

type RootConfigV1Wrapper struct {
	RootConfigV1
}

func (w *RootConfigV1Wrapper) Normalize(sourceFile dt.Filepath, opts cfgstore.Options) (err error) {
	var co *Options
	co, err = dtx.AssertType[*Options](opts)
	if err != nil {
		goto end
	}
	err = w.RootConfigV1.Normalize(sourceFile, co)
end:
	return err
}
func (w *RootConfigV1Wrapper) MarshalJSON() ([]byte, error) {
	return jsonv2.Marshal(w.RootConfigV1)
}

func (w *RootConfigV1Wrapper) UnmarshalJSON(b []byte) error {
	return jsonv2.Unmarshal(b, &w.RootConfigV1)
}
func LoadRootConfigV1(args LoadRootConfigV1Args) (_ *RootConfigV1, err error) {
	var lrc RootConfigV1Wrapper

	configStores := args.ConfigStores
	if configStores == nil {
		configStores = cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
			ConfigStoreArgs: cfgstore.ConfigStoreArgs{
				ConfigSlug:  args.AppInfo.AppSlug(),
				RelFilepath: args.AppInfo.ConfigFile(),
			},
		})
	}

	lrc = RootConfigV1Wrapper{
		RootConfigV1: RootConfigV1{},
	}

	// Get externally set options such as via the switches on the command line
	err = configStores.LoadRootConfig(&lrc, cfgstore.RootConfigArgs{
		DirTypes: args.DirTypes,
		Options:  args.Options,
	})

	return &lrc.RootConfigV1, err
}

func ensureConfig(cs cfgstore.ConfigStore, opts *Options) (rc *RootConfigV1, err error) {
	rc, err = loadConfigIfExists(cs, opts)
	if err != nil {
		// A real error occurred, bail out
		goto end
	}

	if rc == nil {
		// Config not loaded, need to create config
		rc, err = createConfig(cs, opts)
		goto end
	}

end:
	return rc, err
}

func createConfig(cs cfgstore.ConfigStore, opts *Options) (rc *RootConfigV1, err error) {
	var api *APIConfigV2
	var db *SQLite3ConfigV1
	var server *ServerConfigV1
	var fp dt.Filepath

	api = NewAPIConfigV2(DefaultWebroot)

	m := &APIParamsMap{}
	m.Set("q", "string")
	m.Set("sort", "string:enum[asc,desc]")
	m.Set("limit", "int:range[1..50]")
	m.Set("@note", "Just a little bit of info\nfor posterity")

	// Add tasks search endpoint with path parameter and params map
	api.AddEndpoint(NewAPIEndpointV2("GET", "/tasks/search/{project_id:int}", APIEndpointV2Args{
		Description: "Search tasks within a given project (path param project_id + query-string param q)",
		Query:       "SELECT t.id, t.title, t.status, t.priority, IFNULL(au.email,'') AS assignee_email FROM tasks t LEFT JOIN users au ON au.id = t.assignee_id WHERE t.project_id = :project_id AND (LOWER(t.title) LIKE LOWER('%' || :q || '%') OR LOWER(t.details) LIKE LOWER('%' || :q || '%')) ORDER BY t.priority DESC, t.id;",
		Cardinality: string(dbqvars.ManyRows),
		RowType:     string(dbqvars.ColumnsRowType),
		ColumnTypes: []string{
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataType),
		},
		Params: m,
	}))

	// Add tasks by project endpoint with array params
	api.AddEndpoint(NewAPIEndpointV2("GET", "/tasks/by-project/{project:string}", APIEndpointV2Args{
		Description: "Tasks for a project using project in the path and owner email as a query-string parameter",
		Query:       "SELECT t.id, t.title, t.status, t.priority, t.due_date, au.email AS assignee_email, au.name AS assignee_name, t.created_at FROM tasks t JOIN projects p ON p.id = t.project_id JOIN users ou ON ou.id = p.owner_id LEFT JOIN users au ON au.id = t.assignee_id WHERE ou.email = :email AND p.name = :project ORDER BY t.priority DESC, t.created_at;",
		Cardinality: string(dbqvars.ManyRows),
		RowType:     string(dbqvars.ColumnsRowType),
		ColumnTypes: []string{
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataTypeOrNULL),
			string(dbqvars.StringDBDataTypeOrNULL),
			string(dbqvars.StringDBDataTypeOrNULL),
			string(dbqvars.StringDBDataType),
		},
		Params: APIParamsV1{
			{NameSpec: "email", Type: "string"},
		},
	}))

	// Add user by ID endpoint
	api.AddEndpoint(NewAPIEndpointV2("GET", "/users/{id:int}", APIEndpointV2Args{
		Description: "Get a single user by numeric id (path parameter only)",
		Query:       "SELECT id, email, name, created_at FROM users WHERE id = :id;",
		Cardinality: string(dbqvars.OneRow),
		RowType:     string(dbqvars.ColumnsRowType),
		ColumnTypes: []string{
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
		},
	}))

	// Add hello world endpoint
	api.AddEndpoint(NewAPIEndpointV2("GET", "/hello", APIEndpointV2Args{
		Description: "Hello World Endpoint",
		Query:       "SELECT 'Hello World';",
		Cardinality: string(dbqvars.OneRow),
		RowType:     string(dbqvars.StringRowType),
		Params:      APIParamsV1{},
	}))

	// Add projects by owner endpoint
	api.AddEndpoint(NewAPIEndpointV2("GET", "/projects/by-owner/{email:string}", APIEndpointV2Args{
		Description: "Projects owned by a given user (owner email as a path parameter)",
		Query:       "SELECT p.id, p.name, p.status, p.created_at FROM projects p WHERE p.owner_id = (SELECT id FROM users WHERE email = :email) ORDER BY p.created_at DESC;",
		Cardinality: string(dbqvars.ManyRows),
		RowType:     string(dbqvars.ColumnsRowType),
		ColumnTypes: []string{
			string(dbqvars.IntegerDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
			string(dbqvars.StringDBDataType),
		},
	}))

	db = NewSQLite3ConfigV1(DefaultSQLite3Database)
	db.Extensions = nil
	db.OnOpenSQL = []string{"PRAGMA foreign_keys = OFF;"}

	server = NewServerConfigV1(common.DefaultServerHost, ServerConfigV1Args{
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
	err = rc.Normalize(fp, opts)
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

func loadConfigIfExists(cs cfgstore.ConfigStore, opts *Options) (rc *RootConfigV1, err error) {
	var fp dt.Filepath

	if !cs.Exists() {
		goto end
	}

	rc = &RootConfigV1{}
	err = cs.LoadJSON(rc, nil)
	if err != nil {
		goto end
	}
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.Normalize(fp, opts)
end:
	return rc, err
}

func loadAPIFileIfExists(apiFile string) (api *APIConfigV2, err error) {
	var apiBytes []byte
	if apiFile == "" {
		goto end
	}
	apiBytes, err = cfgstore.ReadFileIfExists(apiFile)
	if err != nil {
		err = NewErr(ErrFailedToLoadAPIConfigFile, err)
		goto end
	}
	api = &APIConfigV2{}
	err = jsonv2.Unmarshal(apiBytes, &api)
	if err != nil {
		err = NewErr(ErrFailedToUnmarshalAPIConfigFile, err)
		goto end
	}
end:
	if err != nil {
		err = WithErr(err,
			"api_file", apiFile,
		)
	}
	return api, err
}
