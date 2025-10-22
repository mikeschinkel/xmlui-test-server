package cfgldr

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/dbqvars"
)

const (
	SQLite3ConfigV1Version = 1
	SQLite3ConfigV1Schema  = "https://schemas.xmlui.org/v1/test-server/sqlite3-schema.json"
)

func init() {
	registerDatabaseConfig(&SQLite3ConfigV1{})
}

var _ DatabaseConfig = (*SQLite3ConfigV1)(nil)

type SQLite3ConfigV1 struct {
	Schema         string                      `json:"$schema,omitempty"`
	Version        int                         `json:"version,omitempty"`
	Notes          []string                    `json:"@notes,omitempty"`
	Type           string                      `json:"type"`
	Filepath       string                      `json:"filepath"`
	Extensions     []*SQLite3ExtensionConfigV1 `json:"-"` // `json:"extensions"`
	OnOpenSQL      []string                    `json:"on_open_sql"`
	BusyTimeout    int                         `json:"busy_timeout"`
	JournalMode    string                      `json:"journal_mode"`
	Synchronous    string                      `json:"synchronous"`
	ForeignKeys    string                      `json:"foreign_keys"`
	AutoCheckpoint int                         `json:"wal_autocheckpoint"`
	AccessMode     int                         `json:"access_mode"`
	bootstrapSQL   []string
	sourceFile     string
}

func (c *SQLite3ConfigV1) Clone() DatabaseConfig {
	newCfg := *c

	newCfg.OnOpenSQL = cloneSlice(c.OnOpenSQL)
	newCfg.bootstrapSQL = cloneSlice(c.bootstrapSQL)
	newCfg.Extensions = make([]*SQLite3ExtensionConfigV1, len(c.Extensions))
	for i, ext := range c.Extensions {
		newCfg.Extensions[i] = ext.Clone()
	}

	return &newCfg
}

func (c *SQLite3ConfigV1) SetBootstrapQueries(queries []string) {
	c.bootstrapSQL = queries
}

func (c *SQLite3ConfigV1) BootstrapQueries() []string {
	return c.bootstrapSQL
}

func (c *SQLite3ConfigV1) OnOpenQueries() []string {
	return c.OnOpenSQL
}

func (c *SQLite3ConfigV1) SourceFile() string {
	return c.sourceFile
}

func NewSQLite3ConfigV1(filepath string) *SQLite3ConfigV1 {
	if len(filepath) != 0 && filepath[0] != '/' && filepath[0] != '.' {
		filepath = "./" + filepath
	}
	return &SQLite3ConfigV1{
		Schema:   SQLite3ConfigV1Schema,
		Version:  SQLite3ConfigV1Version,
		Notes:    make([]string, 0),
		Type:     string(SQLite3Database),
		Filepath: filepath,
	}
}
func (c *SQLite3ConfigV1) ConnectString() string {
	return c.Filepath
}

func (c *SQLite3ConfigV1) SetConnectString(cs string) {
	c.Filepath = cs
}

func (c *SQLite3ConfigV1) Port() int {
	return 0
}

func (c *SQLite3ConfigV1) AddDBExtension(ext DBExtensionConfig) {
	sExt, ok := ext.(*SQLite3ExtensionConfigV1)
	if !ok {
		panic(fmt.Sprintf("Cannot type assert a value of type %T to type %T for extension %s",
			ext, (*SQLite3ExtensionConfigV1)(nil),
			ext.ErrorName(),
		))
	}
	c.Extensions = append(c.Extensions, sExt)
}

func (c *SQLite3ConfigV1) DBExtensions() (exts []DBExtensionConfig) {
	exts = make([]DBExtensionConfig, len(c.Extensions))
	for i, ext := range c.Extensions {
		exts[i] = ext
	}
	return exts
}

func (c *SQLite3ConfigV1) DatabaseConfig() {}

func (c *SQLite3ConfigV1) DatabaseType() DatabaseType {
	return SQLite3Database
}

func (c *SQLite3ConfigV1) AddExtension(ext *SQLite3ExtensionConfigV1, opts *Options) (err error) {
	ext.Normalize(common.AppConfigPath, opts)
	c.Extensions = append(c.Extensions, ext)
	return err
}

func (c *SQLite3ConfigV1) SetExtensions(exts []*SQLite3ExtensionConfigV1) {
	c.Extensions = exts
}
func (c *SQLite3ConfigV1) normalizeExtensions(sourceFile string, opts *Options) {
	if len(c.Extensions) == 0 {
		c.Extensions = make([]*SQLite3ExtensionConfigV1, 0)
		goto end
	}
	for _, ext := range c.Extensions {
		ext.Normalize(sourceFile, opts)
	}
end:
	return
}

func (c *SQLite3ConfigV1) Normalize(sourceFile string, opts *Options) {
	c.sourceFile = sourceFile
	if c.Schema == "" {
		c.Schema = SQLite3ConfigV1Schema
	}
	if c.Version == 0 {
		c.Version = SQLite3ConfigV1Version
	}
	if c.Type == "" {
		c.Type = string(SQLite3Database)
	}
	if len(c.bootstrapSQL) == 0 {
		c.bootstrapSQL = make([]string, 0)
	}
	if len(c.OnOpenSQL) == 0 {
		c.OnOpenSQL = make([]string, 0)
	}
	if opts.ConnectString != "" {
		c.SetConnectString(opts.ConnectString)
	}
	if opts.DBAccessMode != int(dbqvars.UnspecifiedDBAccessMode) {
		c.AccessMode = opts.DBAccessMode
	}
	if c.AccessMode == int(dbqvars.UnspecifiedDBAccessMode) {
		c.AccessMode = DefaultDBAccessMode
	}
	c.normalizeExtensions(sourceFile, opts)
	return
}

var _ DBExtensionConfig = (*SQLite3ExtensionConfigV1)(nil)

type SQLite3ExtensionConfigV1 struct {
	Notes        []string          `json:"@notes,omitempty"`
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
	AllowVTable  bool              `json:"allow_vtable"`
	SourceFile   string            `json:"-"`
}

func (c *SQLite3ExtensionConfigV1) ErrorName() string {
	return c.Name
}

func (*SQLite3ExtensionConfigV1) DBExtensionConfig() {}

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
func (c *SQLite3ExtensionConfigV1) Normalize(sourceFile string, opts *Options) {
	var filePath string
	var err error
	c.SourceFile = sourceFile

	switch {
	case c.Filepath != "":
		filePath = c.Filepath
	case len(c.DownloadURLs) != 0:
		cs := cfgstore.NewConfigStoreWithFilename(common.AppConfigPath, c.DownloadURLs[0], cfgstore.DefaultConfigDirType)
		filePath, err = cs.GetFilepath()
		if err != nil {
			cliutil.Errorf("Failed to get full filepath for SQLite3 extension '%s'; %v", c.Name, err)
			logger.Error("Failed to get full SQLite3 extension filepath", "extension", c.Name, "error", err)
		}
		c.Filepath = filePath
	default:
		cliutil.Errorf("You must specify either Filepath or a DownloadURL in SQLite3 Config for extension '%s'; %v", c.Name, err)
		logger.Error("Neither filepath nor DownloadURLs specified in SQLite3 Config for extension", "extension", c.Name, "error", err)
		err = errors.New("")
		goto end
	}
	if c.Id == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		c.Id = base[:len(base)-len(ext)]
	}
	if c.Notes == nil {
		c.Notes = make([]string, 0)
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
	return
}

func (c *SQLite3ExtensionConfigV1) Clone() (ext *SQLite3ExtensionConfigV1) {
	ext = &SQLite3ExtensionConfigV1{}
	*ext = *c
	ext.DownloadURLs = cloneSlice(c.DownloadURLs)
	ext.DependsOn = cloneSlice(c.DependsOn)
	ext.OnLoadSQL = cloneSlice(c.OnLoadSQL)
	ext.EnvVars = cloneMap(c.EnvVars)
	ext.SHA256s = cloneMap(c.SHA256s)
	return ext
}
