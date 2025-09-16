package common

import (
	"time"
)

type Options struct {
	Timeout               time.Duration
	HTTPPort              ServerPort
	DBExtensionFiles      []Filepath
	APIFile               Filepath
	ConnectString         ConnectString
	DBPort                ServerPort
	DBSchemaFile          Filepath
	Quiet                 bool
	AllowUntrustedQueries bool
}
