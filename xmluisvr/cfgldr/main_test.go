package cfgldr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/go-fsfix"
	"github.com/xmlui-org/xmluisvr"
	"github.com/xmlui-org/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmluisvr/common"
	"github.com/xmlui-org/xmluisvr/testutil"
)

const testDataDir = "./test-data"

// TestMain sets up the test environment and runs all integration tests.
func TestMain(m *testing.M) {

	logger := testutil.NewNullLogger()
	xmluisvr.SetLogger(logger) // TODO Change to a buffered logger

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
func setupFixtures(t *testing.T) (rootFix *fsfix.RootFixture, csMap cfgutil.ConfigStoreDirTypeMap) {
	wd, _ := os.Getwd()
	csMap = cfgutil.GetConfigStoreDirTypeMap(common.AppConfigPath, cfgldr.RootConfigFile)
	dotCS := csMap[cfgutil.DefaultConfigDirType]
	localCS := csMap[cfgutil.LocalConfigDir]

	rootFix = fsfix.NewRootFixture("config")

	dotFix := rootFix.AddDirFixture(t, "dot-config", &fsfix.DirFixtureArgs{Parent: rootFix})
	dotFix.AddFileFixture(t, cfgldr.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    cfgldr.RootConfigFile,
		Content: string(testutil.LoadFile(t, filepath.Join(wd, testDataDir, "dot-config."+cfgldr.RootConfigFile), true)),
	})

	localFix := rootFix.AddDirFixture(t, "local-config", &fsfix.DirFixtureArgs{Parent: rootFix})
	localFix.AddFileFixture(t, cfgldr.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    cfgldr.RootConfigFile,
		Content: string(testutil.LoadFile(t, filepath.Join(wd, testDataDir, "local-config."+cfgldr.RootConfigFile), true)),
	})

	rootFix.Create(t)

	localCS.SetConfigDir(localFix.Dir())
	dotCS.SetConfigDir(dotFix.Dir())

	return rootFix, csMap
}
