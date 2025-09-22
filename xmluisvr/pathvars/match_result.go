package pathvars

type VarsMap map[string]string

// MatchResult with private params for memory efficiency
type MatchResult struct {
	Index   int
	varsMap VarsMap
}

func NewMatchResult(index int, varsMap VarsMap) MatchResult {
	return MatchResult{
		Index:   index,
		varsMap: varsMap,
	}
}

// ParamsMap returns the map of parameters
func (m MatchResult) ParamsMap() VarsMap {
	if m.varsMap == nil {
		m.varsMap = make(VarsMap)
	}
	return m.varsMap
}

// GetValue returns a parameter value
func (m MatchResult) GetValue(name string) (value string, found bool) {
	value, found = m.varsMap[name]
	return value, found
}

// VarCount returns the number of parameters
func (m MatchResult) VarCount() int {
	return len(m.varsMap)
}

// HasVars returns true if there are any parameters
func (m MatchResult) HasVars() bool {
	return len(m.varsMap) > 0
}

// ForEachVar iterates over all parameters
func (m MatchResult) ForEachVar(fn func(name, value string) bool) {
	for name, value := range m.varsMap {
		if fn(name, value) {
			continue
		}
		goto end
	}
end:
	return
}
