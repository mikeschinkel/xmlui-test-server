package sqlite3pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

type (
	EntryPoint string
	SQLQuery   string
)

var _ dbpkg.DBExtension = (*Extension)(nil)

type Extension struct {
	id           common.ExtensionId
	version      common.Version
	name         string
	docsURL      common.FullURL
	repoURL      common.FullURL
	downloadURLs []common.FullURL
	filePath     common.Filepath // Absolute or relative filepath, defaults to well-known directory structure
	loadOrder    common.LoadOrder
	entryPoint   EntryPoint
	dependsOn    []DependsOn
	sha256s      map[common.OSArch]common.SHA256
	onFailure    common.OnFailure
	onLoadSQL    []SQLQuery
	envVars      common.EnvironmentVars
	varScope     EnvVarScope // load or app
}

func NewExtension(filePath common.Filepath) *Extension {
	name := filepath.Base(string(filePath))
	name = name[:len(name)-len(filepath.Ext(name))]
	return &Extension{
		id:           common.ExtensionId(name),
		version:      cfgldr.UnknownVersion,
		name:         name,
		downloadURLs: make([]common.FullURL, 0),
		filePath:     filePath,
		loadOrder:    0,
		entryPoint:   cfgldr.DefaultSQLite3ExtensionEntryPoint,
		dependsOn:    make([]DependsOn, 0),
		sha256s:      make(map[common.OSArch]common.SHA256),
		onFailure:    cfgldr.DefaultOnFailurePolicy,
		onLoadSQL:    make([]SQLQuery, 0),
		envVars:      make(common.EnvironmentVars),
		varScope:     cfgldr.DefaultVarScope,
	}
}

type ExtensionOptions struct {
	// Dev escape hatch: if true, also enable SQL `load_extension()` globally.
	AllowPragmaOverload bool
}

func (ext *Extension) Load(conn *sqlite3.SQLiteConn, opts ExtensionOptions) (err error) {
	var path string
	// Important: do NOT globally enable load_extension via SQL unless explicitly requested.
	path, err = ext.Resolve(opts)
	if err != nil || path == "" {
		return err
	}
	err = conn.LoadExtension(path, "")
	if err != nil {
		return fmt.Errorf("LoadExtension(%s): %w", path, err)
	}
	// Optional: enable arbitrary SQL `load_extension()` only if explicitly requested.
	if opts.AllowPragmaOverload {
		// Not recommended; requires C call. Keep this off by default.
		// conn.EnableLoadExtension(true) doesn't exist; LoadExtension already toggles on/off per call.
	}
	return err
}

func (ext *Extension) Name() string {
	return ext.name
}

func (ext *Extension) DBExtension() {}

// Resolve decides which file to load and where it came from.
func (ext *Extension) Resolve(opts ExtensionOptions) (path string, err error) {
	return "", nil
}

func configRoot() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "xmlui", "sqlite", "exts"), nil
}

func platformKey() string { return runtime.GOOS + "-" + runtime.GOARCH }

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// findUserVersions returns available versions in config dir for an extension.
func findUserVersions(name string) ([]string, error) {
	root, err := configRoot()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(root, name)
	d, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var vs []string
	for _, e := range d {
		if e.IsDir() {
			vs = append(vs, e.Name())
		}
	}
	sort.Strings(vs) // lexical; OK if you use simple semver
	return vs, nil
}

func verifySHA256(path, want string) error {
	if want == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	got := sha256Hex(b)
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("sha256 mismatch: got %s want %s", got, want)
	}
	return nil
}

func extDir(name, version string) (string, error) {
	root, err := configRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, name, version, platformKey()), nil
}
