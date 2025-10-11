package dbpkg

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

type QuerySource struct {
	Lines    [2]int
	Source   common.QueryString
	Filepath common.Filepath
}

func NewQuerySource(start, end int, src common.QueryString, fp common.Filepath) *QuerySource {
	return &QuerySource{
		Lines:    [2]int{start, end},
		Source:   src,
		Filepath: fp,
	}
}
