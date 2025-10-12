package cfgstore_test

import (
	"log"
	"os"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	logger, err := testutil.GetBufferedLogger()
	if err != nil {
		log.Fatalf("Failed to get buffered logger: %v", err)
	}

	cfgstore.SetLogger(logger)

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
