package dbpkg

import (
	"github.com/mikeschinkel/go-dt"
	"github.com/xmlui-org/localsvr/xmluisvr/common"
)

type QuerySource struct {
	Lines    [2]int
	Source   common.QueryString
	Filepath dt.Filepath
}

func NewQuerySource(start, end int, src common.QueryString, fp dt.Filepath) *QuerySource {
	return &QuerySource{
		Lines:    [2]int{start, end},
		Source:   src,
		Filepath: fp,
	}
}
