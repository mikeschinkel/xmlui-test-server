// Package common provides shared utilities, types, and constants used throughout xmlui-test-server.
//
// This package contains foundational utilities that are used across all other packages:
//
//   - Type definitions for file paths, identifiers, URLs, and other common concepts
//   - Parsing and validation functions for user input and configuration
//   - HTTP method and header utilities
//   - Database data type definitions and parsing
//   - Error definitions and handling utilities
//   - Logging setup and management
//   - File system operations and path manipulation
//
// # Core Types
//
// The package defines several string-based types for type safety:
//   - Filepath: File system paths (absolute or relative)
//   - URLPath: HTTP URL paths with validation
//   - Identifier: Safe identifiers for variables and parameters
//   - ConnectString: Database connection strings
//   - HTTPMethod: HTTP request methods with validation
//
// # Parsing Functions
//
// Most types have corresponding ParseBytes* functions that validate input and return
// typed values or descriptive errors:
//
//	port, err := common.ParseServerPort(8080, common.ZeroInvalid)
//	method, err := common.ParseHTTPMethod("GET", common.EmptyInvalid)
//	path, err := common.ParseURLPath("/api/v1/users")
//
// # Error Handling
//
// The package defines common error types and provides utilities for error
// handling throughout the application, following Go best practices for
// error wrapping and context.
package common

const (
	// AppName is the human-readable name of the application.
	AppName = "XMLUI Local Server"

	// DefaultServerPort is the default HTTP port when none is specified.
	DefaultServerPort = 8080

	// LocalHostIP is the IP address for localhost connections.
	LocalHostIP = "127.0.0.1"

	// DefaultServerHost is the default host address for the HTTP server.
	DefaultServerHost = LocalHostIP

	// AppConfigPath is the default path for application configuration files.
	AppConfigPath = "xmlui"

	RootConfigFile = "test-server.json"
)

const (
	// DefaultCardinality specifies the default expected row count for database queries.
	DefaultCardinality = ManyRowsOrNone

	// DefaultRowType specifies the default format for returning database query results.
	DefaultRowType = ColumnsRowType

	// DefaultDBDataType specifies the default data type for database columns when none is specified.
	DefaultDBDataType = StringDBDataType
)
