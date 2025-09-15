package cfgutil_test

import (
	"os"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	common.SetLogger(testutil.NullLogger())

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
