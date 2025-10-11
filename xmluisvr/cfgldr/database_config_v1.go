package cfgldr

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
)

type DatabaseConfig interface {
	DatabaseConfig()       // Marker
	Clone() DatabaseConfig // Marker
	DatabaseType() DatabaseType
	ConnectString() string
	SetConnectString(string)
	Port() int
	BootstrapQueries() []string
	SetBootstrapQueries([]string)
	OnOpenQueries() []string
	SourceFile() string
	DBExtensions() []DBExtensionConfig
	Normalize(string, *Options) // TODO MAYBE Change to accept an any parameter
}

type DBExtensionConfig interface {
	DBExtensionConfig()
	ErrorName() string // Name to display in error messages when not able to type assert
}

var _ DBExtensionConfig = (*dbExtensionConfig)(nil)

type dbExtensionConfig struct {
	filePath string
}

func (d dbExtensionConfig) ErrorName() string {
	return cfgstore.GetBaseFilename(d.filePath)
}

func NewDBExtensionConfig(file string) DBExtensionConfig {
	return &dbExtensionConfig{
		filePath: file,
	}
}

func (d dbExtensionConfig) DBExtensionConfig() {}
