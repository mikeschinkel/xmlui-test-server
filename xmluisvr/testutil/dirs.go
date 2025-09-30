package testutil

import (
	"testing"

	"github.com/mikeschinkel/go-fsfix"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// SetupConfigDirFixtures sets up a root fixture with two dir fixtures, one to
// emulate the user's ~/.config/xmlui config directory and the other to emulate
// the project's ./.xmlui config directory. The userFile and projectFile should
// be filenames containing hte respective config files for each.
func SetupConfigDirFixtures(t *testing.T, testDataDir, userFile, projectFile string) (rootFix *fsfix.RootFixture, csMap cfgutil.ConfigStoresMap) {
	const (
		userDir    = ".config"
		projectDir = "project"
	)
	csMap = cfgutil.GetConfigStoresMap(common.AppConfigPath, cfgldr.RootConfigFile)
	dotCS := csMap[cfgutil.DefaultConfigDirType]
	localCS := csMap[cfgutil.LocalConfigDir]

	rootFix = fsfix.NewRootFixture("config")

	dotFix := rootFix.AddDirFixture(t, userDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	dotFix.AddFileFixture(t, cfgldr.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    cfgldr.RootConfigFile,
		Content: string(LoadFile(t, userFile, true)),
	})

	localFix := rootFix.AddDirFixture(t, projectDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	localFix.AddFileFixture(t, cfgldr.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    cfgldr.RootConfigFile,
		Content: string(LoadFile(t, projectFile, true)),
	})

	rootFix.Create(t)

	localCS.SetConfigDir(localFix.Dir())
	dotCS.SetConfigDir(dotFix.Dir())

	return rootFix, csMap
}
