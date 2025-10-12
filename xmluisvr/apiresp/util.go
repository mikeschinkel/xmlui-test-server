package apiresp

import (
	"encoding/json/jsontext"
	"fmt"
	"os"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// Format JSON is a pretty manner
func prettifyJSON(responseJSON []byte) (prettyJSON jsontext.Value) {
	var err error

	prettyJSON = responseJSON
	err = prettyJSON.Indent(jsontext.WithIndent("  "))
	if err != nil {
		cliutil.Errorf("Error prettifying JSON for logging: %v", err)
		cliutil.Errorf("Raw response: %s", string(responseJSON))
		goto end
	}
end:
	return prettyJSON
}

func stderrf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprint(os.Stderr, msg)
	common.Logger().Error(msg)
}
