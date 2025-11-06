package common

import (
	"time"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
)

var _ interface{ Options() } = (*Options)(nil)

type Options struct {
	Timeout               time.Duration
	HTTPPort              ServerPort
	DBExtensionFiles      []dt.Filepath
	APIFile               dt.Filepath
	ConnectString         ConnectString
	DBPort                ServerPort
	DBBootstrapFile       dt.Filepath
	Verbosity             cliutil.Verbosity
	ErrorStyle            ErrorStyle
	AllowUntrustedQueries bool
	Quiet                 bool
}

func (Options) Options() {}
