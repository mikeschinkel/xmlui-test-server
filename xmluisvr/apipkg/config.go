package apipkg

// Config represents a configuration interface for API components.
// This interface is used to provide a consistent way to access
// configuration settings across the API package.
type Config interface {
	// Config returns the configuration details.
	Config()
}
