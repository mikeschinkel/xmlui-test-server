package xmluisvr

type QueryRequest struct {
	SQL    string `json:"sql"`
	Params []any  `json:"params"`
}

type EndpointDefinition struct {
	Path    string                      `json:"path"`
	Methods map[string]MethodDefinition `json:"methods"`
}

type MethodDefinition struct {
	Description string   `json:"description"`
	SQL         string   `json:"sql,omitempty"`
	SQLFile     string   `json:"sqlFile,omitempty"`
	Params      []string `json:"params,omitempty"`
}
