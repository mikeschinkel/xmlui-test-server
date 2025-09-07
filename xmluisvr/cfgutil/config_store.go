package cfgutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// DefaultConfigDirType is currently hardcoded for ~/.config but having this
// const will make it easy to track down how where to change it if we want to make it
// configurable.
const DefaultConfigDirType = DotConfigDir

const ConfigBaseDirName = ".config"

// ConfigStore provides file operations for Gmail API
type ConfigStore interface {
	Load() ([]byte, error)
	Save([]byte) error
	LoadJSON(data any) error
	SaveJSON(data any) error
	Exists() bool
	GetFilepath() (string, error)
	SetFilename(string)
	SetConfigDir(string)
	ConfigDir() (string, error)
	ConfigStore()
}

var _ ConfigStore = (*configStore)(nil)

type configStore struct {
	appPath   string
	configDir string
	filename  string
	dirType   ConfigDirType
	fs        fs.FS
}

func (s *configStore) ConfigStore() {}

func (s *configStore) Create(appConfigDirPath, filename string, dirType ConfigDirType) ConfigStore {
	cs := NewConfigStore(appConfigDirPath, dirType)
	cs.SetFilename(filename)
	return cs
}

type ConfigDirType int

const (
	UnspecifiedConfigDir ConfigDirType = iota
	DotConfigDir                       // ~/.config/xmlui
	LocalConfigDir                     // ./.xmlui
	GoUserConfigDir                    // The value os.UserConfigDir() returns
)

type ConfigStoreOpts struct {
	DirType ConfigDirType
}

func NewConfigStore(appConfigDirPath string, dirType ConfigDirType) ConfigStore {
	return NewConfigStoreWithFilename(appConfigDirPath, "", dirType)
}

func NewConfigStoreWithFilename(appConfigDirPath, filename string, dirType ConfigDirType) ConfigStore {
	return &configStore{
		appPath:  appConfigDirPath,
		filename: filename,
		dirType:  dirType,
	}
}

func (s *configStore) ConfigDir() (dir string, err error) {

	if s.configDir != "" {
		goto end
	}

	switch s.dirType {
	case DotConfigDir:
		dir, err = os.UserHomeDir()
		if err != nil {
			goto end
		}
		s.configDir = filepath.Join(dir, ConfigBaseDirName, s.appPath)
	case LocalConfigDir:
		dir, err = os.Getwd()
		if err != nil {
			goto end
		}
		s.configDir = filepath.Join(dir, "."+s.appPath)
	case GoUserConfigDir:
		dir, err = os.UserConfigDir()
		if err != nil {
			goto end
		}
		s.configDir = filepath.Join(dir, s.appPath)
	case UnspecifiedConfigDir:
		err = fmt.Errorf("config dir type not set")
	default:
		err = fmt.Errorf("invalid config dir type: %d", s.dirType)
	}

end:
	return s.configDir, err
}

func (s *configStore) getFS() (_ fs.FS, err error) {
	var dir string

	if s.fs != nil {
		goto end
	}

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	s.fs = os.DirFS(dir)

end:
	return s.fs, err
}

func (s *configStore) ensureFilepath() (fp string, err error) {
	fp, err = s.GetFilepath()
	// This is needed in case filename contains a subdirectory, e.g. tokens/token-bill@microsoft.com.json
	err = os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		goto end
	}
end:
	return fp, err
}

func (s *configStore) SetFilename(fn string) {
	s.filename = fn
}

func (s *configStore) GetFilepath() (fp string, err error) {
	var dir string

	dir, err = s.ConfigDir()
	if err != nil {
		goto end
	}

	if !fs.ValidPath(s.filename) {
		err = fmt.Errorf("path %s is not valid for use in %s", s.filename, dir)
		goto end
	}

	fp = filepath.Join(s.configDir, s.filename)

end:
	return fp, err
}

func (s *configStore) Save(data []byte) (err error) {
	var file *os.File
	var fullPath string

	ensureLogger()

	fullPath, err = s.ensureFilepath()
	if err != nil {
		goto end
	}

	file, err = os.Create(fullPath)
	if err != nil {
		goto end
	}
	defer mustClose(file)

	_, err = file.Write(data)

end:
	return err
}

func (s *configStore) SaveJSON(data any) (err error) {
	var jsonData []byte

	jsonData, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		goto end
	}

	err = s.Save(jsonData)

end:
	return err
}

func (s *configStore) Load() (data []byte, err error) {
	var fSys fs.FS

	fSys, err = s.getFS()
	if err != nil {
		err = errors.Join(ErrFailedToGetConfigFileSystem, err)
		goto end
	}

	data, err = fs.ReadFile(fSys, s.filename)
	if err != nil {
		err = errors.Join(ErrFailedToReadFile, err)
		goto end
	}

end:
	return data, err
}

func (s *configStore) LoadJSON(data any) (err error) {
	var jsonData []byte

	jsonData, err = s.Load()
	if err != nil {
		err = errors.Join(ErrFailedToReadConfigFile, err)
		goto end
	}

	err = json.Unmarshal(jsonData, data)
	if err != nil {
		err = errors.Join(ErrFailedToUnmarshalConfigFile, err)
		goto end
	}

end:
	return err
}

func (s *configStore) Exists() (exists bool) {
	fSys, err := s.getFS()
	if err != nil {
		goto end
	}
	_, err = fs.Stat(fSys, s.filename)
	exists = err == nil

end:
	return exists
}

// SetConfigDir allows overriding config dir for unit testing.
func (s *configStore) SetConfigDir(dir string) {
	s.configDir = dir
	s.fs = os.DirFS(dir)
}

type ConfigStoreDirTypeMap = map[ConfigDirType]ConfigStore

func GetConfigStoreDirTypeMap(appConfigDirPath, file string) ConfigStoreDirTypeMap {
	return map[ConfigDirType]ConfigStore{
		DefaultConfigDirType: NewConfigStoreWithFilename(appConfigDirPath, file, DefaultConfigDirType),
		LocalConfigDir:       NewConfigStoreWithFilename(appConfigDirPath, file, LocalConfigDir),
	}
}
