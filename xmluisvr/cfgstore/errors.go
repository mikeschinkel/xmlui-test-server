package cfgstore

import (
	"errors"
)

var (
	ErrFailedToGetConfigFileSystem = errors.New("failed to get config file system")
	ErrFailedToReadFile            = errors.New("failed to read file")
	ErrFailedToReadConfigFile      = errors.New("failed to read config file")
	ErrFailedToUnmarshalConfigFile = errors.New("failed to unmarshal config file")
	ErrFileDoesNotExist            = errors.New("file does not exist")
)
