package cfgldr_test

import (
	"testing"

	"github.com/mikeschinkel/go-jsontest"
	"github.com/stretchr/testify/require"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/fsutil"

	_ "github.com/mikeschinkel/go-jsontest/pipefuncs"
)

func TestCreateGoldenData(t *testing.T) {
	api := cfgldr.NewAPIConfigV2(".")
	api.AddEndpoint(cfgldr.NewAPIEndpointV2("GET /hello", cfgldr.APIEndpointV2Args{
		Description: "Hello World Endpoint",
		Query:       "SELECT 'Hello World';",
		Cardinality: "one",
		RowType:     "string",
	}))
	db := cfgldr.NewSQLite3ConfigV1("data.db")
	err := db.AddExtension(&cfgldr.SQLite3ExtensionConfigV1{
		Filepath: "steampipe_sqlite_github.so",
	})
	if err != nil {
		t.Error(err.Error())
	}
	server := cfgldr.NewServerConfigV1(common.LocalHostIP, cfgldr.ServerConfigV1Args{
		Port: 8080,
		API:  api,
	})
	root := cfgldr.NewRootConfigV1(cfgldr.RootConfigV1Args{
		ServerConfig: server,
		DBConfig:     db,
	})
	err = fsutil.WriteJSONFile("./test-data/test-server.json", root, 0644, 0755)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestLoadRootConfigV1(t *testing.T) {
	type args struct {
		appName string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "LoadJSON Config",
			args: args{},
			want: map[string]any{
				"$schema":                              "https://schemas.xmlui.org/v1/test-server-root-schema.json",
				"$schemaVersion":                       1,
				"server|exists()":                      true,
				"server.$schema":                       "https://schemas.xmlui.org/v1/test-server-server-schema.json",
				"server.$schemaVersion":                1,
				"server.host":                          "127.0.0.1",
				"server.port":                          8080,
				"server.api|exists()":                  true,
				"server.api.$schema":                   "https://schemas.xmlui.org/v2/test-server-api-schema.json",
				"server.api.$schemaVersion":            2,
				"server.api.name":                      "User-definable XMLUI Local Server APIConfig",
				"server.api.base_path":                 "/api",
				"server.api.webroot":                   "./webroot",
				"server.api.endpoints|exists()":        true,
				"server.api.endpoints.0.endpoint":      "GET /hello",
				"server.api.endpoints.0.description":   "Hello World Endpoint",
				"server.api.endpoints.0.query":         "SELECT 'Hello World';",
				"server.api.endpoints.0.query_file":    "",
				"server.api.endpoints.0.params":        "{}",
				"server.api.endpoints.0.rows_expected": "one",
				"server.api.endpoints.0.row_type":      "string",
				"server.api.endpoints.0.column_types":  []string{},
				"database|exists()":                    true,
				"database.$schema":                     "https://schemas.xmlui.org/v1/test-server-sqlite3-schema.json",
				"database.$schemaVersion":              1,
				"database.type":                        "sqlite3",
				"database.filepath":                    "./data/data.db",
				"database.extensions|exists()":         true,
				"database.extensions.0.id":             "steampipe_sqlite_github",
				"database.extensions.0.version":        "v0.0.0",
				"database.extensions.0.name":           "steampipe_sqlite_github",
				"database.extensions.0.docs_url":       "",
				"database.extensions.0.repo_url":       "",
				"database.extensions.0.download_urls":  []string{},
				"database.extensions.0.filepath":       "steampipe_sqlite_github.so",
				"database.extensions.0.load_order":     0,
				"database.extensions.0.entry_point":    "sqlite3_extension_init",
				"database.extensions.0.depends_on":     []string{},
				"database.extensions.0.sha256s":        "{}",
				"database.extensions.0.on_failure":     "warn",
				"database.extensions.0.post_load_sql":  []string{},
				"database.extensions.0.env_vars":       "{}",
				"database.extensions.0.vars_scope":     "app",
				"database.init_sql":                    "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootFix, csMap := setupFixtures(t)
			defer rootFix.Cleanup()
			gotRc, err := cfgldr.LoadRootConfigV1FromConfigStoreMap(csMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRootConfigV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = jsontest.TestJSON(gotRc.Bytes(), tt.want)
			require.NoError(t, err, "Config did not match expected values")
		})
	}
}

//func TestNewRootConfigV1(t *testing.T) {
//	tests := []struct {
//		name string
//		want *RootConfigV1
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := NewRootConfigV1(); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("NewRootConfigV1() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestRootConfigV1_Config(t *testing.T) {
//	type fields struct {
//		Schema        string
//		SchemaVersion int
//		Server        *ServerConfigV1
//		Database      *DatabaseConfigV1
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &RootConfigV1{
//				Schema:        tt.fields.Schema,
//				SchemaVersion: tt.fields.SchemaVersion,
//				Server:        tt.fields.Server,
//				Database:      tt.fields.Database,
//			}
//			c.Config()
//		})
//	}
//}
