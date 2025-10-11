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
	DBBootstrapFile       Filepath
	Verbosity             Verbosity
	ErrorStyle            ErrorStyle
	AllowUntrustedQueries bool
}
