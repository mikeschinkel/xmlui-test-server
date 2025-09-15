package cfgutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type testData struct {
	Name string
	Age  int
}

func getConfigStore(filename string, dirType cfgutil.ConfigDirType) cfgutil.ConfigStore {
	return cfgutil.NewConfigStoreWithFilename("test-app", filename, dirType).(cfgutil.ConfigStore)
}

func TestConfigStore_SaveLoadExists(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "xmlui-test-"+uuid.NewString())
	t.Cleanup(func() {
		common.LogOnError(os.RemoveAll(dir))
	})

	filename := "config/testdata.json"
	cs := getConfigStore(filename, cfgutil.DefaultConfigDirType)
	cs.SetConfigDir(dir)

	data := testData{Name: "Alice", Age: 42}

	err = cs.SaveJSON(&data)
	require.NoError(t, err)

	exists := cs.Exists()
	assert.True(t, exists)

	var loaded testData
	err = cs.LoadJSON(&loaded)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestConfigStore_LoadNonexistent(t *testing.T) {
	var err error

	cs := getConfigStore("does-not-exist.json", cfgutil.DefaultConfigDirType)
	cs.SetConfigDir(t.TempDir())

	err = cs.LoadJSON(&testData{})
	assert.Error(t, err)
}

func TestConfigStore_SaveInvalidJSON(t *testing.T) {
	var err error

	cs := getConfigStore("bad.json", cfgutil.DefaultConfigDirType)
	cs.SetConfigDir(t.TempDir())

	ch := make(chan int) // non-serializable
	err = cs.SaveJSON(ch)
	assert.Error(t, err)
}

func TestConfigStore_ConfigDir(t *testing.T) {
	cs := getConfigStore("", cfgutil.DefaultConfigDirType)
	dir := t.TempDir()

	cs.SetConfigDir(dir)

	cfgDir, err := cs.ConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, dir, cfgDir)
}
