package apipkg

import (
	"context"
)

// Context is an alias for context.Context, providing request scoping and cancellation
// throughout the API package operations.
type Context = context.Context
