// Package test provides comprehensive integration tests for XMLUI Local Server.
//
// This package contains end-to-end tests that validate server functionality.
package test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"
)

var (
	// testDataDir is the shared test data directory for all integration tests
	testDataDir string
	// bootstrapSQL contains the shared bootstrap SQL content
	bootstrapSQL string
)

// TestMain sets up the test environment and runs all integration tests.
func TestMain(m *testing.M) {
	// Logger is already set up in init() function above

	// This ensures the logger is set up before cfgldr package initialization
	logger := slog.New(testutil.NewBufferedLogHandler())
	common.SetLogger(logger)

	// Setup common test data that all tests can use
	wd, _ := os.Getwd()
	testDataDir = filepath.Join(wd, "test-data")

	// Pre-load the bootstrap SQL that all tests will use
	bootstrapSQL = string(testutil.LoadFile(nil, filepath.Join(testDataDir, "bootstrap_test.sql"), true))

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}
