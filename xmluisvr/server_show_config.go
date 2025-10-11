package xmluisvr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

func (svr *Server) showConfig() {

	cliutil.Printf("\n")
	cliutil.Printf("Configuration:\n")
	cliutil.Printf("- Current Dir:  %v\n", svr.displayDir())
	cliutil.Printf("- Server:       %s\n", svr.Host())
	cliutil.Printf("  - Webroot:    %v\n", svr.displayWebroot())
	cliutil.Printf("  - Config:     %s\n", svr.displayServerSourceFile())
	cliutil.Printf("- Database:     %s\n", svr.displayDBTypeName())
	cliutil.Printf("  - Connection: %v\n", svr.displayDBName())
	cliutil.Printf("  - Config:     %s\n", svr.displayDBSourceFile())
	if svr.api != nil {
		cliutil.Printf("- API:          %s\n", svr.api.Name)
		cliutil.Printf("  - URL Path:   %s\n", svr.api.BasePath)
		cliutil.Printf("  - Config:     %s\n", svr.displayAPISourceFile())
	}
	if len(svr.db.Extensions()) != 0 {
		cliutil.Printf("- Extension:   %s\n", svr.displayExtensionPaths())
	}
	cliutil.Loud().Printf("\n")

}

func (svr *Server) displayHost() string {
	return fmt.Sprintf("localhost:%d", svr.port)
}

func (svr *Server) displayWebroot() (wr string) {
	// Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		wr = err.Error()
		goto end
	}
	if wd == "" {
		wr = "Working directory unavailable"
		goto end
	}
	wd = filepath.Join(wd, string(svr.api.Webroot))
	wr = common.HomeRelative(wd)
end:
	return wr
}

func (svr *Server) displayAPISourceFile() string {
	return common.HomeRelative(string(svr.api.SourceFile))
}
func (svr *Server) displayDBSourceFile() string {
	return common.HomeRelative(string(svr.db.SourceFile()))
}
func (svr *Server) displayServerSourceFile() string {
	return common.HomeRelative(string(svr.sourceFile))
}

func (svr *Server) displayDir() (d string) {
	// Print current working directory
	wd, err := os.Getwd()
	if err != nil {
		d = err.Error()
		goto end
	}
	if wd == "" {
		d = "Working directory unavailable"
		goto end
	}
	d = common.HomeRelative(wd)
end:
	return d
}

func (svr *Server) displayDBName() (name string) {
	if svr.db == nil {
		return ""
	}
	return svr.db.String()
}
func (svr *Server) displayDBTypeName() (name string) {
	return svr.db.TypeName()
}

func (svr *Server) displayExtensionPaths() (paths string) {
	var sb strings.Builder
	for _, ext := range svr.db.Extensions() {
		sb.WriteString(ext.Name())
		sb.WriteString(". ")
	}
	paths = sb.String()
	if len(paths) != 0 {
		paths = paths[:len(paths)-2]
	}
	return paths
}
