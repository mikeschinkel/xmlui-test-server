package apiutil

import (
	"fmt"
	"os"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

func stderrf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprint(os.Stderr, msg)
	common.Logger().Error(msg)
}
