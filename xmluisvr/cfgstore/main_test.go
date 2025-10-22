package cfgstore_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"
)

var bufferedLog *testutil.BufferedLogHandler

func TestMain(m *testing.M) {
	var logger *slog.Logger
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	logger, bufferedLog = testutil.GetBufferedLogger()

	cfgstore.SetLogger(logger)

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
