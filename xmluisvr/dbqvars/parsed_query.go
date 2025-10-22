package dbqvars

type ParsedQuery interface {
	QueryString() QueryString
	Parameters() Parameters
	Occurrences() QueryTokens
}
