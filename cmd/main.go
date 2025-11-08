// Package main/cmd provides the entry point for the xmlui-test-server CLI application.
//
// This is a lightweight HTTP server that provides:
//   - Static file serving from the current directory
//   - /query endpoint for executing SQL against SQLite or PostgreSQL databases
//   - /proxy endpoint for proxying requests to external APIs (CORS bypass)
//   - Optional SQLite extension loading (e.g., Steampipe plugins)
//   - API endpoint routing based on JSON configuration files
package main

import (
	"github.com/xmlui-org/localsvr/xmluisvr"
)

// main is the application entry point that starts the xmlui-test-server CLI.
func main() {
	xmluisvr.RunCLI(nil)
}
