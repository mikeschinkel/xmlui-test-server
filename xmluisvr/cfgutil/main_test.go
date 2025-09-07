package cfgutil_test

import (
	"os"
	"testing"

	"github.com/xmlui-org/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmluisvr/testutil"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	cfgutil.SetLogger(testutil.NullLogger())

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
