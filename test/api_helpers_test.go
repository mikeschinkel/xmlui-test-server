package test

// api_helpers_test.go does not contain tests but is named with an _test.go
// suffix so Goland will stop flagging BootstrapSQL() and DataDir() as missing
// for some godforsaken reason.

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
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgstore"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/rfc9457"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/testutil"
)

// EnvironmentName identifies the type of test environment configuration.
// This determines the directory name for test fixtures and provides context
// for debugging test failures.
type EnvironmentName string

const (
	// ComprehensiveTestEnv uses api_comprehensive_test.json which contains all API
	// endpoints for integration testing. This config includes:
	//   - All data type validations (int, string, uuid, slug, boolean, real, date, alphanumeric)
	//   - All constraint types (range, length, enum, regex, notempty, format)
	//   - Query parameters and implicit type inference
	//   - Multi-segment parameters
	//   - Various response formats (cardinality, row_type)
	// Used by: api_datatypes_test.go, api_constraints_test.go, api_parameters_test.go, etc.
	ComprehensiveTestEnv EnvironmentName = "comprehensive"

	// CustomTestEnv indicates a test-specific configuration passed directly to setupTestServer.
	// Use this when you need a specialized API config that differs from the comprehensive one.
	// Example: Testing a single endpoint with specific edge cases.
	CustomTestEnv EnvironmentName = "custom"
)

// testRequest represents a single HTTP request test case
type testRequest struct {
	name            string
	method          string
	path            string
	body            string
	expectedStatus  int
	expectedFields  []string          // Fields that should exist in response (for success cases)
	shouldContain   []string          // Content that should be in response (legacy, prefer expectedRFC9457)
	shouldNotHave   []string          // Content that should NOT be in response
	expectedRFC9457 *rfc9457.Response // Expected RFC 9457 error response (for error cases)
}

// testEnvironment holds the test environment setup for a single test case
type testEnvironment struct {
	rootFixture        *fsfix.RootFixture
	dbPath             string
	bootstrapFile      *fsfix.FileFixture
	configFile         *fsfix.FileFixture
	configStoreMap     cfgstore.ConfigStoresMap
	bufferedLogHandler *testutil.BufferedLogHandler
	bufferedWriter     *testutil.BufferedWriter
	logger             *slog.Logger
}

// testServer represents a running test server instance
type testServer struct {
	BaseURL     string
	Port        int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
	serverError error
	env         *testEnvironment
	t           *testing.T
}

// =============================================================================
// RFC 9457 Assertion Helper
// =============================================================================

// assertRFC9457Equal compares two Response structs field by field
// with clear error messages for each difference.
func assertRFC9457Equal(t *testing.T, got, want *rfc9457.Response) {
	t.Helper()

	if got == nil && want == nil {
		return
	}
	if got == nil {
		t.Error("Got nil Response, want non-nil")
		return
	}
	if want == nil {
		t.Error("Want nil Response, got non-nil")
		return
	}

	// Verify all RFC 9457 fields exhaustively
	if got.Type != want.Type {
		t.Errorf("Type: got '%s', want '%s'", got.Type, want.Type)
	}
	if got.Title != want.Title {
		t.Errorf("Title: got '%s', want '%s'", got.Title, want.Title)
	}
	if got.Status != want.Status {
		t.Errorf("Status: got %d, want %d", got.Status, want.Status)
	}
	if got.Detail != want.Detail {
		t.Errorf("Detail: got '%s', want '%s'", got.Detail, want.Detail)
	}
	if got.Instance != want.Instance {
		t.Errorf("Instance: got '%s', want '%s'", got.Instance, want.Instance)
	}
	if got.Parameter != want.Parameter {
		t.Errorf("Parameter: got '%s', want '%s'", got.Parameter, want.Parameter)
	}
	if got.ExpectedType != want.ExpectedType {
		t.Errorf("ExpectedType: got '%s', want '%s'", got.ExpectedType, want.ExpectedType)
	}
	if got.ReceivedValue != want.ReceivedValue {
		t.Errorf("ReceivedValue: got '%s', want '%s'", got.ReceivedValue, want.ReceivedValue)
	}
	if got.Location != want.Location {
		t.Errorf("Location: got '%s', want '%s'", got.Location, want.Location)
	}
	if got.Suggestion != want.Suggestion {
		t.Errorf("Suggestion: got '%s', want '%s'", got.Suggestion, want.Suggestion)
	}

	// Compare Constraint (handles nil and any type)
	if want.Constraint != nil {
		if got.Constraint != want.Constraint {
			t.Errorf("Constraint: got '%v', want '%v'", got.Constraint, want.Constraint)
		}
	} else if got.Constraint != nil {
		t.Errorf("Constraint: got '%v', want nil", got.Constraint)
	}

	// Compare ValidationErrors slice
	if len(got.ValidationErrors) != len(want.ValidationErrors) {
		t.Errorf("ValidationErrors length: got %d, want %d", len(got.ValidationErrors), len(want.ValidationErrors))
	} else {
		for i := range want.ValidationErrors {
			gotErr := got.ValidationErrors[i]
			wantErr := want.ValidationErrors[i]
			if gotErr.Parameter != wantErr.Parameter {
				t.Errorf("ValidationErrors[%d].Parameter: got '%s', want '%s'", i, gotErr.Parameter, wantErr.Parameter)
			}
			if gotErr.Location != wantErr.Location {
				t.Errorf("ValidationErrors[%d].Location: got '%s', want '%s'", i, gotErr.Location, wantErr.Location)
			}
			if gotErr.Expected != wantErr.Expected {
				t.Errorf("ValidationErrors[%d].Expected: got '%s', want '%s'", i, gotErr.Expected, wantErr.Expected)
			}
			if gotErr.Received != wantErr.Received {
				t.Errorf("ValidationErrors[%d].Received: got '%s', want '%s'", i, gotErr.Received, wantErr.Received)
			}
			if gotErr.Message != wantErr.Message {
				t.Errorf("ValidationErrors[%d].Message: got '%s', want '%s'", i, gotErr.Message, wantErr.Message)
			}
		}
	}
}

// =============================================================================
// Test Environment Setup
// =============================================================================

// setupTestEnvironment creates an isolated test environment for a single test case
func setupTestEnvironment(t *testing.T, envName EnvironmentName, configContent string) *testEnvironment {
	t.Helper()

	// Create ONE root fixture for this test environment
	rootFix := fsfix.NewRootFixture(string(envName))

	// Get the config stores map
	csMap := cfgstore.GetConfigStoresMap(common.AppConfigPath, common.RootConfigFile)
	dotCS := csMap[cfgstore.DefaultConfigDirType]
	localCS := csMap[cfgstore.LocalConfigDir]

	// Create .config directory for user config
	dotFix := rootFix.AddDirFixture(t, ".config", &fsfix.DirFixtureArgs{Parent: rootFix})
	configFile := dotFix.AddFileFixture(t, common.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    common.RootConfigFile,
		Content: configContent,
	})

	// Create project directory for project config (same content for simplicity)
	localFix := rootFix.AddDirFixture(t, "project", &fsfix.DirFixtureArgs{Parent: rootFix})
	localFix.AddFileFixture(t, common.RootConfigFile, &fsfix.FileFixtureArgs{
		Name:    common.RootConfigFile,
		Content: configContent,
	})

	// Create the fixture directory structure on disk
	rootFix.Create(t)

	// Configure the config stores to use our fixture directories
	localCS.SetConfigDir(localFix.Dir())
	dotCS.SetConfigDir(dotFix.Dir())

	// Create bootstrap file in the test data directory
	bootstrapFile := &fsfix.FileFixture{
		Name:     "bootstrap.sql",
		Content:  BootstrapSQL(),
		Filepath: filepath.Join(DataDir(), "bootstrap.sql"),
	}

	// Generate unique database path
	dbPath := filepath.Join(rootFix.Dir(), "test.db")

	// Setup buffered logger and CLI writer
	logger, handler := testutil.GetBufferedLogger()
	bufferedWriter := testutil.NewBufferedWriter()

	// Set as global logger
	common.SetLogger(logger)

	return &testEnvironment{
		rootFixture:        rootFix,
		dbPath:             dbPath,
		bootstrapFile:      bootstrapFile,
		configFile:         configFile,
		configStoreMap:     csMap,
		bufferedLogHandler: handler,
		bufferedWriter:     bufferedWriter,
		logger:             logger,
	}
}

// cleanup cleans up the test environment
func (env *testEnvironment) cleanup(t *testing.T) {
	t.Helper()
	env.rootFixture.Cleanup()
	// Remove database file if it exists
	if _, err := os.Stat(env.dbPath); err == nil {
		if err := os.Remove(env.dbPath); err != nil {
			t.Logf("Warning: failed to remove database file: %v", err)
		}
	}
}

// =============================================================================
// Test Server Management
// =============================================================================

// setupTestServer creates and starts a test server instance
func setupTestServer(t *testing.T, envName EnvironmentName, configContent string) *testServer {
	t.Helper()

	env := setupTestEnvironment(t, envName, configContent)

	// Find available port
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}

	// Configure options
	options := &cfgldr.Options{
		HTTPPort:        port,
		ConnectString:   env.dbPath,
		DBBootstrapFile: env.bootstrapFile.Filepath,
		Timeout:         300,
		Verbosity:       3, // Max verbosity for debugging
	}

	// Load root config
	rootConfig, err := cfgldr.LoadRootConfigV1FromConfigStoreMap(env.configStoreMap, options)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Setup run arguments
	runArgs := &xmluisvr.RunArgs{
		CLIArgs:   []string{},
		Options:   options,
		Config:    rootConfig,
		CLIWriter: env.bufferedWriter,
		Logger:    env.logger,
	}

	// Start server in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	var serverError error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		serverError = xmluisvr.Run(ctx, runArgs)
		if serverError != nil {
			t.Errorf("Server error: %v", serverError)
		}
	}()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Wait for server to be ready by polling health endpoint
	if !waitForServerReady(t, baseURL, 10*time.Second) {
		cancel()
		wg.Wait()
		if serverError != nil {
			t.Errorf("Server failed with error: %v", serverError)
		}
		// Print logs to help diagnose
		logEntries, _ := env.bufferedLogHandler.GetLogEntries()
		if len(logEntries) > 0 {
			t.Errorf("Server logs (%d entries):", len(logEntries))
			for i, entry := range logEntries {
				t.Errorf("  [%d] %v", i+1, entry)
			}
		}
		t.Logf("Server output:\n%s", env.bufferedWriter.GetAllOutput())
		t.Fatalf("Server did not become ready within timeout")
	}

	server := &testServer{
		BaseURL:     baseURL,
		Port:        port,
		ctx:         ctx,
		cancel:      cancel,
		wg:          &wg,
		serverError: serverError,
		env:         env,
		t:           t,
	}

	return server
}

// Cleanup shuts down the server and cleans up resources
func (s *testServer) Cleanup() {
	s.t.Helper()

	// Stop server
	s.cancel()

	// Wait for graceful shutdown
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Server shut down gracefully
	case <-time.After(5 * time.Second):
		s.t.Log("Warning: Server did not shut down gracefully within 5 seconds")
	}

	// Check for server errors (ignore context cancellation)
	if s.serverError != nil && !strings.Contains(s.serverError.Error(), "context canceled") {
		s.t.Errorf("Server error: %v", s.serverError)
	}

	// Cleanup environment
	s.env.cleanup(s.t)
}

// =============================================================================
// Test Request Execution
// =============================================================================

// runTestRequest executes a test request against the server and validates the response
func (s *testServer) runTestRequest(req testRequest) {
	s.t.Helper()

	s.t.Run(req.name, func(t *testing.T) {
		t.Helper()

		// Log error context helper
		onErr := func() {
			logEntries, _ := s.env.bufferedLogHandler.GetLogEntries()
			t.Logf("Log entries: %v", logEntries)
			t.Log(s.env.bufferedWriter.GetAllOutput())
		}

		// Make HTTP request
		resp, respBody, err := makeHTTPRequest(s.BaseURL, req.method, req.path, req.body)
		if err != nil {
			onErr()
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer closeOrError(t, resp.Body)

		responseBody := string(respBody)

		// Verify status code
		if resp.StatusCode != req.expectedStatus {
			onErr()
			logEntries, err := s.env.bufferedLogHandler.GetLogEntries()
			if err != nil {
				t.Logf("Error getting log entries: %v", err)
			} else if len(logEntries) > 0 {
				t.Logf("Buffered log entries (%d total):", len(logEntries))
				for i, entry := range logEntries {
					t.Logf("  [%d] %v", i+1, entry)
				}
			}
			t.Errorf("Expected status %d, got %d. Response: %s",
				req.expectedStatus, resp.StatusCode, responseBody)
			return
		}

		// Verify RFC 9457 error response if expected
		if req.expectedRFC9457 != nil {
			var got9457 rfc9457.Response
			if err := json.Unmarshal(respBody, &got9457); err != nil {
				t.Errorf("Response is not valid RFC 9457 JSON: %v. Body: %s", err, responseBody)
				return
			}
			assertRFC9457Equal(t, &got9457, req.expectedRFC9457)
		}

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

		// Legacy validation (to be removed after migration to expectedRFC9457)
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
	})
}

// makeHTTPRequest creates and executes an HTTP request
func makeHTTPRequest(baseURL, method, path, body string) (*http.Response, []byte, error) {
	url := baseURL + path

	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	httpReq, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		common.CloseOrLog(resp.Body)
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, respBody, nil
}

// =============================================================================
// Utility Functions
// =============================================================================

// findAvailablePort finds an available port for testing
func findAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer common.LogOnError(listen.Close())

	port := listen.Addr().(*net.TCPAddr).Port
	return port, nil
}

// waitForServerReady polls the server's /api/hello endpoint until it responds or timeout occurs
func waitForServerReady(t *testing.T, baseURL string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	healthURL := baseURL + "/api/hello"

	for time.Now().Before(deadline) {
		resp, err := http.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// getRootConfig wraps an API endpoint definition in a full root config structure
func getRootConfig(endpointJSON string) string {
	return fmt.Sprintf(`{
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
					%s
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
	}`, endpointJSON)
}

// closeOrError closes an io.Closer and fails the test if there's an error
func closeOrError(t *testing.T, closer io.Closer) {
	t.Helper()
	if err := closer.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}
}

// setupComprehensiveTestServer sets up a test server using the comprehensive test config.
// This is a convenience wrapper that loads api_comprehensive_test.json which contains
// all API endpoints for integration testing (data types, constraints, parameters, etc.).
func setupComprehensiveTestServer(t *testing.T) *testServer {
	t.Helper()

	// Read the comprehensive test config that has all endpoints defined
	configBytes, err := os.ReadFile("./test-data/api_comprehensive_test.json")
	if err != nil {
		t.Fatalf("Failed to read comprehensive test config: %v", err)
	}

	// Use the ComprehensiveTestEnv constant for type safety and documentation
	return setupTestServer(t, ComprehensiveTestEnv, string(configBytes))
}

// runTestRequest executes a test request and validates the response
func runTestRequest(t *testing.T, baseURL string, req testRequest) {
	t.Helper()

	resp, body, err := makeHTTPRequest(baseURL, req.method, req.path, req.body)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer closeOrError(t, resp.Body)

	// Check status code
	if resp.StatusCode != req.expectedStatus {
		t.Errorf("Status code: got %d, want %d\nBody: %s", resp.StatusCode, req.expectedStatus, string(body))
	}

	// If we expect an RFC9457 error response, validate it
	if req.expectedRFC9457 != nil {
		var got rfc9457.Response
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("Failed to unmarshal RFC9457 response: %v\nBody: %s", err, string(body))
		}
		assertRFC9457Equal(t, &got, req.expectedRFC9457)
		return
	}

	// Check expected fields in JSON response
	if len(req.expectedFields) > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("Failed to unmarshal JSON response: %v\nBody: %s", err, string(body))
		}
		for _, field := range req.expectedFields {
			if _, exists := result[field]; !exists {
				t.Errorf("Expected field '%s' not found in response: %s", field, string(body))
			}
		}
	}

	// Check shouldContain (legacy)
	bodyStr := string(body)
	for _, contains := range req.shouldContain {
		if !strings.Contains(bodyStr, contains) {
			t.Errorf("Response should contain '%s' but doesn't.\nBody: %s", contains, bodyStr)
		}
	}

	// Check shouldNotHave
	for _, notHave := range req.shouldNotHave {
		if strings.Contains(bodyStr, notHave) {
			t.Errorf("Response should NOT contain '%s' but does.\nBody: %s", notHave, bodyStr)
		}
	}
}
