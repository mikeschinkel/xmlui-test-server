package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mikeschinkel/go-fsfix"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgutil"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"

	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/apipkg"
	_ "github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
)

// testRequest represents a single HTTP request test case
type testRequest struct {
	name           string
	method         string
	path           string
	body           string
	expectedStatus int
	expectedFields []string // Fields that should exist in response
	shouldContain  []string // Content that should be in response
	shouldNotHave  []string // Content that should NOT be in response
}

// TestAPIEndpoints_FullIntegration tests API endpoint permutations using xmluisvr.Run()
// This is a comprehensive integration test covering PathVars parameter types and constraints
func TestAPIEndpoints_FullIntegration(t *testing.T) {
	// Test cases covering key PathVars permutations - starting with a few, can expand later
	tests := []struct {
		name          string
		configContent string          // JSON config content for this test case
		options       *cfgldr.Options // Options to pass to xmluisvr.Run()
		testRequests  []testRequest   // HTTP requests to test
	}{
		{
			name: "basic_endpoints_sqlite",
			configContent: `{
				"$schema": "https://schemas.xmlui.org/v1/test-server/root-schema.json",
				"version": 1,
				"server": {
					"$schema": "https://schemas.xmlui.org/v1/test-server/server-schema.json",
					"version": 1,
					"host": "127.0.0.1",
					"port": 8080,
					"api": {
						"$schema": "https://schemas.xmlui.org/v2/test-server/api-schema.json",
						"version": 2,
						"name": "Test XMLUI Local Server API",
						"base_path": "/api",
						"webroot": ".",
						"endpoints": [
							{
								"method": "GET",
								"path": "hello",
								"description": "Hello World Endpoint",
								"query": "SELECT 'Hello World';",
								"query_file": "",
								"cardinality": "one",
								"row_type": "string",
								"column_types": [],
								"params": {}
							},
							{
								"method": "GET",
								"path": "users/{id:int}",
								"description": "Get user by ID",
								"query": "SELECT id, email, name, slug FROM users WHERE id = {id};",
								"query_file": "",
								"cardinality": "one",
								"row_type": "columns",
								"column_types": ["integer", "string", "string", "string"],
								"params": {"id": "int"}
							}
						]
					}
				},
				"database": {
					"$schema": "https://schemas.xmlui.org/v1/test-server/sqlite3-schema.json",
					"version": 1,
					"type": "sqlite3",
					"filepath": "test.db",
					"on_open_sql": [],
					"busy_timeout": 0,
					"journal_mode": "",
					"synchronous": "",
					"foreign_keys": "",
					"wal_autocheckpoint": 0
				}
			}`,
			testRequests: []testRequest{
				//{
				//	name:           "hello_endpoint",
				//	method:         "GET",
				//	path:           "/api/hello",
				//	expectedStatus: 200,
				//	expectedFields: []string{"message"},
				//	shouldContain:  []string{"Hello World"},
				//},
				{
					name:           "user_by_id_basic_int",
					method:         "GET",
					path:           "/api/users/1",
					expectedStatus: 200,
					expectedFields: []string{"id", "email", "name", "slug"},
					shouldContain:  []string{"alice@example.com", "Alice Carter"},
				},
			},
		},
		// TODO: Add more test cases for different parameter types and constraints
		// - String parameters with constraints
		// - Real/float parameters
		// - Boolean parameters
		// - Date parameters
		// - UUID parameters
		// - Multi-segment parameters
		// - Query parameters with defaults
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test environment using helper function
			env := setupTestEnvironment(t, tc.name, tc.configContent)
			defer env.cleanup(t)

			// Find an available port for testing
			availablePort, err := findAvailablePort()
			if err != nil {
				t.Fatalf("Failed to find available port: %v", err)
			}

			// Configure options for xmluisvr.Run() - override with our test values
			options := &cfgldr.Options{
				HTTPPort:        availablePort, // Dynamic port for testing
				ConnectString:   env.dbPath,
				DBBootstrapFile: env.bootstrapFile.Filepath,
				// No APIFile - the API is embedded in the main config
				Timeout:   300,
				Verbosity: 3, // Max verbosity for debugging
			}

			// Load root config using proper config store map (this triggers bootstrap SQL loading)
			rootConfig, err := cfgldr.LoadRootConfigV1FromConfigStoreMap(env.configStoreMap, options)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			//t.Logf("Test config: HTTPPort=%d, ConnectString=%s, DBBootstrapFile=%s",options.HTTPPort, options.ConnectString, options.DBBootstrapFile)

			// Setup run arguments with proper cfgldr objects
			runArgs := &xmluisvr.RunArgs{
				CLIArgs:   []string{},
				Options:   options,
				Config:    rootConfig,
				CLIWriter: env.bufferedWriter,
				Logger:    env.logger,
			}

			// Start the server in a goroutine using xmluisvr.Run()
			serverCtx, serverCancel := context.WithCancel(context.Background())

			var serverError error
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				//t.Logf("Starting xmluisvr.Run() in goroutine...")
				// THIS IS THE ACTUAL xmluisvr.Run() CALL!
				serverError = xmluisvr.Run(serverCtx, runArgs)
				//t.Logf("xmluisvr.Run() finished with error: %v", serverError)
			}()

			// Wait for server to start up
			//t.Logf("Waiting for server to start...")
			time.Sleep(2 * time.Second) // Give it more time to start

			// Test the actual HTTP endpoints
			baseURL := fmt.Sprintf("http://127.0.0.1:%d", options.HTTPPort)

			for _, req := range tc.testRequests {
				t.Run(req.name, func(t *testing.T) {
					// Create HTTP request URL
					url := baseURL + req.path
					//t.Logf("Making %s request to: %s", req.method, url)

					// Create HTTP request
					var reqBody io.Reader
					if req.body != "" {
						reqBody = strings.NewReader(req.body)
					}

					httpReq, err := http.NewRequest(req.method, url, reqBody)
					if err != nil {
						t.Fatalf("Failed to create request: %v", err)
					}

					if req.body != "" {
						httpReq.Header.Set("Content-Type", "application/json")
					}

					// Execute request
					client := &http.Client{Timeout: 300 * time.Second}
					resp, err := client.Do(httpReq)
					if err != nil {
						t.Fatalf("Failed to execute request: %v", err)
					}
					defer closeOrError(t, resp.Body)

					// Read response body
					respBody, err := io.ReadAll(resp.Body)
					if err != nil {
						t.Fatalf("Failed to read response body: %v", err)
					}

					// Verify status code
					if resp.StatusCode != req.expectedStatus {
						// Output buffered logs for debugging
						logEntries, err := env.bufferedLogHandler.GetLogEntries()
						if err != nil {
							//t.Logf("Error getting log entries: %v", err)
						} else if len(logEntries) > 0 {
							//t.Logf("Buffered log entries (%d total):", len(logEntries))
							//for i, entry := range logEntries {
							//t.Logf("  [%d] %v", i+1, entry)
							//}
						}
						t.Errorf("Expected status %d, got %d. Response: %s",
							req.expectedStatus, resp.StatusCode, string(respBody))
						return
					}

					responseBody := string(respBody)

					// For successful responses, verify JSON structure
					if resp.StatusCode == 200 && len(req.expectedFields) > 0 {
						var response interface{}
						if err := json.Unmarshal(respBody, &response); err != nil {
							t.Errorf("Response is not valid JSON: %v. Body: %s", err, responseBody)
							return
						}

						// Verify expected fields exist in response
						for _, field := range req.expectedFields {
							if !strings.Contains(responseBody, fmt.Sprintf(`"%s"`, field)) {
								t.Errorf("Expected field '%s' in response, got: %s", field, responseBody)
							}
						}
					}

					// Verify expected content
					for _, expected := range req.shouldContain {
						if !strings.Contains(responseBody, expected) {
							t.Errorf("Expected response to contain '%s', got: %s", expected, responseBody)
						}
					}

					// Verify content that should NOT be present
					for _, notExpected := range req.shouldNotHave {
						if strings.Contains(responseBody, notExpected) {
							t.Errorf("Expected response to NOT contain '%s', got: %s", notExpected, responseBody)
						}
					}

					//t.Logf("✓ HTTP %s %s returned %d as expected", req.method, req.path, resp.StatusCode)
				})
			}

			// Stop the server and wait for it to finish
			serverCancel()

			// Give the server a moment to shut down gracefully
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Server shut down gracefully
			case <-time.After(5 * time.Second):
				// Server didn't shut down in time
				//t.Logf("Server did not shut down gracefully within 5 seconds")
			}

			// Check if server had any errors (ignore context cancellation)
			if serverError != nil && !strings.Contains(serverError.Error(), "context canceled") {
				t.Errorf("Server error: %v", serverError)
			}

			//t.Logf("✓ Successfully tested xmluisvr.Run() with %d HTTP requests for test case: %s", len(tc.testRequests), tc.name)
		})
	}
}

func closeOrError(t *testing.T, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}
}

// testEnvironment holds the test environment setup for a single test case
type testEnvironment struct {
	rootFixture        *fsfix.RootFixture
	dbPath             string
	bootstrapFile      *fsfix.FileFixture
	configFile         *fsfix.FileFixture
	configStoreMap     cfgutil.ConfigStoresMap
	bufferedLogHandler *testutil.BufferedLogHandler
	bufferedWriter     *testutil.BufferedWriter
	logger             *slog.Logger
}

// setupTestEnvironment creates an isolated test environment for a single test case
func setupTestEnvironment(t *testing.T, testName string, configContent string) *testEnvironment {
	// Create temporary directories for config files
	tempDir := t.TempDir()
	userConfigFile := filepath.Join(tempDir, "user-config.test-server.json")
	projectConfigFile := filepath.Join(tempDir, "project-config.test-server.json")

	// Write the config content to both user and project config files
	// (using the same content for simplicity - could be split if needed)
	err := os.WriteFile(userConfigFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write user config file: %v", err)
	}
	err = os.WriteFile(projectConfigFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write project config file: %v", err)
	}

	// Setup config directory fixtures using testutil (this enables proper bootstrap SQL loading)
	rootFix, configStoreMap := testutil.SetupConfigDirFixtures(t, testDataDir, userConfigFile, projectConfigFile)

	// Create bootstrap file in the test data directory
	bootstrapFile := &fsfix.FileFixture{}
	bootstrapFile.Name = "bootstrap.sql"
	bootstrapFile.Content = bootstrapSQL
	bootstrapFile.Filepath = filepath.Join(testDataDir, "bootstrap_test.sql")

	// Generate unique database path in the fixtures directory
	dbPath := filepath.Join(rootFix.Dir(), "test.db")

	// Setup buffered logger and CLI writer for testing
	logger, handler := testutil.GetBufferedLogger()
	bufferedWriter := testutil.NewBufferedWriter()

	// Set the logger as the global common logger (like xmluisvr/logger.go does)
	common.SetLogger(logger)

	return &testEnvironment{
		rootFixture:        rootFix,
		dbPath:             dbPath,
		bootstrapFile:      bootstrapFile,
		configFile:         &fsfix.FileFixture{Filepath: userConfigFile},
		configStoreMap:     configStoreMap,
		bufferedLogHandler: handler,
		bufferedWriter:     bufferedWriter,
		logger:             logger,
	}
}

// cleanup cleans up the test environment
func (env *testEnvironment) cleanup(t *testing.T) {
	env.rootFixture.Cleanup()
	// Remove database file if it exists
	if _, err := os.Stat(env.dbPath); err == nil {
		reportOnError(t, os.Remove(env.dbPath))
	}
}

func reportOnError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

// findAvailablePort finds an available port for testing
func findAvailablePort() (port int, err error) {
	var addr *net.TCPAddr
	var listen *net.TCPListener

	addr, err = net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		goto end
	}

	listen, err = net.ListenTCP("tcp", addr)
	if err != nil {
		goto end
	}
	defer common.LogOnError(listen.Close())
	port = listen.Addr().(*net.TCPAddr).Port
end:
	return port, err
}
