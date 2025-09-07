package sqlite3pkg

type EnvVarScope string

const (
	LoadScope EnvVarScope = "load" // Scope: Only during extension load
	AppScope  EnvVarScope = "app"  // Scope: Entire app lifetime
)
