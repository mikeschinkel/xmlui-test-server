// Package xmluisvr provides the core HTTP server implementation for xmlui-test-server.
//
// This package implements a lightweight HTTP server that serves multiple purposes:
//
//   - Static file serving from the current directory with index.html fallback
//   - SQL query execution endpoint (/query) supporting both SQLite and PostgreSQL
//   - HTTP proxy endpoint (/proxy) for bypassing CORS restrictions
//   - Configurable API endpoints based on JSON configuration files
//   - Database extension loading for SQLite (e.g., Steampipe plugins)
//
// The server is designed to be used as both a standalone CLI application and
// as a library for embedding in other Go applications.
//
// # Usage as CLI Application
//
// The most common usage is through the CLI:
//
//	./xmlui-test-server --port 3000 --api config.json
//
// # Usage as Library
//
// For embedding in other applications:
//
//	import "github.com/xmlui-org/localdev/xmluisvr"
//
//	ctx := context.Background()
//	args := &xmluisvr.RunArgs{
//		Options:   options,
//		Config:    config,
//		CLIWriter: writer,
//		Logger:    logger,
//	}
//	err := xmluisvr.Run(ctx, args)
//
// # Architecture
//
// The package follows a layered architecture:
//   - Server: Core HTTP server with routing and middleware
//   - Database: Abstracted database layer supporting multiple engines
//   - API: Configuration-driven API endpoint system
//   - Configuration: Hierarchical JSON-based configuration system
package xmluisvr
