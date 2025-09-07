package cfgldr

type APIEndpointV2 struct {
	Endpoint     string            `json:"endpoint"`
	Description  string            `json:"description"`
	Query        string            `json:"query"`
	QueryFile    string            `json:"query_file"`
	Params       map[string]string `json:"params"`
	RowsExpected string            `json:"rows_expected"` // 'one' or 'many'
	RowType      string            `json:"row_type"`      // 'int','real','string','json','columns'
	ColumnTypes  []string          `json:"column_types"`  // used when row_type="columns"
	method       string
	path         string
}

type APIEndpointV2Args struct {
	Description  string
	Query        string
	QueryFile    string
	Params       map[string]string
	RowsExpected string
	RowType      string
	ColumnTypes  []string
}

func NewAPIEndpointV2(endpoint string, args APIEndpointV2Args) *APIEndpointV2 {
	if args.Params == nil {
		args.Params = make(map[string]string)
	}
	if args.ColumnTypes == nil {
		args.ColumnTypes = make([]string, 0)
	}
	return &APIEndpointV2{
		Endpoint:     endpoint,
		Description:  args.Description,
		Query:        args.Query,
		QueryFile:    args.QueryFile,
		Params:       args.Params,
		RowsExpected: args.RowsExpected,
		RowType:      args.RowType,
		ColumnTypes:  args.ColumnTypes,
	}
}
