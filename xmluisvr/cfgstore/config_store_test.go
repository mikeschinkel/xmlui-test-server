package cfgstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
)

type testData struct {
	Name string
	Age  int
}

func getConfigStore(filename string, dirType cfgstore.ConfigDirType) cfgstore.ConfigStore {
	return cfgstore.NewConfigStoreWithFilename("test-app", filename, dirType).(cfgstore.ConfigStore)
}

func TestConfigStore_SaveLoadExists(t *testing.T) {
	var err error
	dir := filepath.Join(os.TempDir(), "xmlui-test-"+uuid.NewString())
	t.Cleanup(func() {
		cfgstore.LogOnError(os.RemoveAll(dir))
	})

	filename := "config/testdata.json"
	cs := getConfigStore(filename, cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(dir)

	data := testData{Name: "Alice", Age: 42}

	err = cs.SaveJSON(&data)
	require.NoError(t, err)

	exists := cs.Exists()
	assert.True(t, exists)

	var loaded testData
	err = cs.LoadJSON(&loaded, nil)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestConfigStore_LoadNonexistent(t *testing.T) {
	var err error

	cs := getConfigStore("does-not-exist.json", cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(t.TempDir())

	err = cs.LoadJSON(&testData{}, nil)
	assert.Error(t, err)
}

func TestConfigStore_SaveInvalidJSON(t *testing.T) {
	var err error

	cs := getConfigStore("bad.json", cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(t.TempDir())

	ch := make(chan int) // non-serializable
	err = cs.SaveJSON(ch)
	assert.Error(t, err)
}

func TestConfigStore_ConfigDir(t *testing.T) {
	cs := getConfigStore("", cfgstore.DefaultConfigDirType)
	dir := t.TempDir()

	cs.SetConfigDir(dir)

	cfgDir, err := cs.ConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, dir, cfgDir)
}
