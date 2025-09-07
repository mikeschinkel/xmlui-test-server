package cfgldr

import (
	"errors"
	"path/filepath"

	"github.com/xmlui-org/xmluisvr/cfgutil"
)

const (
	SQLite3ConfigV1SchemaVersion = 1
	SQLite3ConfigV1Schema        = "https://schemas.xmlui.org/v1/test-server-sqlite3-schema.json"
)

func init() {
	registerDatabaseConfig(&SQLite3ConfigV1{})
}

var _ DatabaseConfig = (*SQLite3ConfigV1)(nil)

type SQLite3ConfigV1 struct {
	Schema            string                      `json:"$schema,omitempty"`
	SchemaVersion     int                         `json:"$schemaVersion,omitempty"`
	Type              string                      `json:"type"`
	Filepath          string                      `json:"filepath"`
	Sqlite3Extensions []*SQLite3ExtensionConfigV1 `json:"extensions"`
	InitSQL           []string                    `json:"init_sql"`
}

func NewSQLite3ConfigV1(filepath string) *SQLite3ConfigV1 {
	if len(filepath) != 0 && filepath[0] != '/' && filepath[0] != '.' {
		filepath = "./" + filepath
	}
	return &SQLite3ConfigV1{
		Schema:        SQLite3ConfigV1Schema,
		SchemaVersion: SQLite3ConfigV1SchemaVersion,
		Type:          string(Sqlite3Database),
		Filepath:      filepath,
	}
}
func (c *SQLite3ConfigV1) ConnectString() string {
	return c.Filepath
}

func (c *SQLite3ConfigV1) Port() int {
	return 0
}

func (c *SQLite3ConfigV1) Extensions() (exts []string) {
	exts = make([]string, len(c.Sqlite3Extensions))
	for i, ext := range c.Sqlite3Extensions {
		exts[i] = ext.Filepath
	}
	return exts
}

func (c *SQLite3ConfigV1) DatabaseConfig() {}

func (c *SQLite3ConfigV1) DatabaseType() DatabaseType {
	return Sqlite3Database
}

func (c *SQLite3ConfigV1) AddExtension(appConfigPath string, ext *SQLite3ExtensionConfigV1) (err error) {
	err = ext.Normalize(appConfigPath)
	if err != nil {
		goto end
	}
	c.Sqlite3Extensions = append(c.Sqlite3Extensions, ext)
end:
	return err
}

func (c *SQLite3ConfigV1) SetExtensions(exts []*SQLite3ExtensionConfigV1) {
	c.Sqlite3Extensions = exts
}

type SQLite3ExtensionConfigV1 struct {
	Id           string            `json:"id"`
	Version      string            `json:"version"`
	Name         string            `json:"name"`
	DocsURL      string            `json:"docs_url"`
	RepoURL      string            `json:"repo_url"`
	DownloadURLs []string          `json:"download_urls"`
	Filepath     string            `json:"filepath"` // Absolute or relative filepath, defaults to well-known directory structure
	LoadOrder    int               `json:"load_order"`
	EntryPoint   string            `json:"entry_point"`
	DependsOn    []string          `json:"depends_on"`
	SHA256s      map[string]string `json:"sha256s"`
	OnFailure    string            `json:"on_failure"` // 'error','warn','ignore'
	OnLoadSQL    []string          `json:"post_load_sql"`
	EnvVars      map[string]string `json:"env_vars"`
	VarScope     string            `json:"vars_scope"` // 'load' or 'app'
}

func (c *SQLite3ExtensionConfigV1) AddDownloadURL(url string) {
	c.DownloadURLs = append(c.DownloadURLs, url)
}
func (c *SQLite3ExtensionConfigV1) AddOnLoadSQL(sql string) {
	c.OnLoadSQL = append(c.OnLoadSQL, sql)
}
func (c *SQLite3ExtensionConfigV1) AddDependsOn(do string) {
	c.DependsOn = append(c.DependsOn, do)
}
func (c *SQLite3ExtensionConfigV1) AddSHA256(name, value string) {
	c.SHA256s[name] = value
}
func (c *SQLite3ExtensionConfigV1) AddEnvVar(name, value string) {
	c.EnvVars[name] = value
}
func (c *SQLite3ExtensionConfigV1) Normalize(appName string) (err error) {
	var filePath string

	switch {
	case c.Filepath != "":
		filePath = c.Filepath
	case c.DownloadURLs != nil:
		filePath = c.DownloadURLs[0]
		cs := cfgutil.NewConfigStoreWithFilename(appName, filePath, cfgutil.DefaultConfigDirType)
		c.Filepath, err = cs.GetFilepath()
		if err != nil {
			goto end
		}
	default:
		err = errors.New("must specify either Filepath or DownloadURLs for SQLite3 Config")
		goto end
	}
	if c.Id == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		c.Id = base[:len(base)-len(ext)]
	}
	if c.Name == "" {
		c.Name = c.Id
	}
	if c.Version == "" {
		c.Version = UnknownVersion
	}
	if c.EntryPoint == "" {
		c.EntryPoint = DefaultSQLite3ExtensionEntryPoint
	}
	if c.OnFailure == "" {
		c.OnFailure = DefaultOnFailurePolicy
	}
	if c.VarScope == "" {
		c.VarScope = DefaultVarScope
	}
	if c.DownloadURLs == nil {
		c.DownloadURLs = make([]string, 0)
	}
	if c.DependsOn == nil {
		c.DependsOn = make([]string, 0)
	}
	if c.OnLoadSQL == nil {
		c.OnLoadSQL = make([]string, 0)
	}
	if c.SHA256s == nil {
		c.SHA256s = make(map[string]string)
	}
	if c.EnvVars == nil {
		c.EnvVars = make(map[string]string)
	}
end:
	return err
}
