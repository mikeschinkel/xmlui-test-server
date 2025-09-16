package xmluisvr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/fsutil"
)

func (s *Server) showConfig() {

	cliutil.Printf("\n")
	cliutil.Printf("Configuration:\n")
	cliutil.Printf("- Current Dir:  %v\n", s.displayDir())
	cliutil.Printf("- Server:       %s\n", s.Host())
	cliutil.Printf("  - Webroot:    %v\n", s.displayWebroot())
	cliutil.Printf("  - Config:     %s\n", s.displayServerSourceFile())
	cliutil.Printf("- Database:     %s\n", s.displayDBTypeName())
	cliutil.Printf("  - Connection: %v\n", s.displayDBName())
	cliutil.Printf("  - Config:     %s\n", s.displayDBSourceFile())
	if s.api != nil {
		cliutil.Printf("- API:          %s\n", s.api.Name)
		cliutil.Printf("  - URL Path:   %s\n", s.api.BasePath)
		cliutil.Printf("  - Config:     %s\n", s.displayAPISourceFile())
	}
	if len(s.db.Extensions()) != 0 {
		cliutil.Printf("- Extension:   %s\n", s.displayExtensionPaths())
	}
	cliutil.Loud().Printf("\n")

}

func (s *Server) displayHost() string {
	return fmt.Sprintf("localhost:%d", s.port)
}

func (s *Server) displayWebroot() (wr string) {
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
	wd = filepath.Join(wd, string(s.api.Webroot))
	wr = fsutil.HomeRelative(wd)
end:
	return wr
}

func (s *Server) displayAPISourceFile() string {
	return fsutil.HomeRelative(string(s.api.SourceFile))
}
func (s *Server) displayDBSourceFile() string {
	return fsutil.HomeRelative(string(s.db.SourceFile()))
}
func (s *Server) displayServerSourceFile() string {
	return fsutil.HomeRelative(string(s.sourceFile))
}

func (s *Server) displayDir() (d string) {
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
	d = fsutil.HomeRelative(wd)
end:
	return d
}

func (s *Server) displayDBName() (name string) {
	if s.db == nil {
		return ""
	}
	return s.db.String()
}
func (s *Server) displayDBTypeName() (name string) {
	return s.db.TypeName()
}

func (s *Server) displayExtensionPaths() (paths string) {
	var sb strings.Builder
	for _, ext := range s.db.Extensions() {
		sb.WriteString(ext.Name())
		sb.WriteString(". ")
	}
	paths = sb.String()
	if len(paths) != 0 {
		paths = paths[:len(paths)-2]
	}
	return paths
}
