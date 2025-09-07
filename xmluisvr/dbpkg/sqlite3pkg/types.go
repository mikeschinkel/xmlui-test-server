package sqlite3pkg

type ResolvedExt struct {
	Name      string
	Path      string
	OnLoadSQL []string
}

type HookCfg struct {
	ScratchSchema string // e.g., "extension_mem"; empty => no attach
	BusyTimeoutMS int    // e.g., 5000
	UseWAL        bool   // true
	Synchronous   string // "NORMAL" or "FULL"
	Extensions    []ResolvedExt
}
