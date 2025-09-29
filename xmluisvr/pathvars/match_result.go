// Package pathvars/match_result defines the MatchResult type which contains
// the results of matching an HTTP request against a route template.
// It provides access to extracted parameter values and route information.
package pathvars

import (
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/common"
)

// VarsMap is a map of parameter names to their extracted string values.
type VarsMap map[common.Identifier]any

func (vm VarsMap) GetValues(namesIn []common.Identifier) (values []any, namesOut []common.Identifier) {
	n := len(namesIn)
	values = make([]any, n)
	namesOut = make([]common.Identifier, n)

	i := 0
	for _, name := range namesIn {
		value, ok := vm[name]
		if !ok {
			continue
		}
		values[i] = value
		namesOut[i] = name
		i++
	}
	return values[:i], namesOut[:i]
}

// MatchResult represents the result of matching an HTTP request against a route template.
// It contains the matched route index and extracted parameter values for memory efficiency.
type MatchResult struct {
	// Index indicates which route was matched in the router's route list.
	Index int

	Route *Route

	// varsMap contains the extracted parameter values from the matched request.
	// This field is private to control access and ensure proper initialization.
	varsMap VarsMap
}

// NewMatchResult creates a new MatchResult with the specified route index and parameter values.
func NewMatchResult(r *Route, varsMap VarsMap) MatchResult {
	return MatchResult{
		Index:   r.Index,
		Route:   r,
		varsMap: varsMap,
	}
}

func (m MatchResult) GetValues(names []common.Identifier) ([]any, []common.Identifier) {
	return m.varsMap.GetValues(names)
}

// ParamsMap returns the map of extracted parameter values.
// If the internal map is nil, it initializes an empty map to prevent nil pointer issues.
func (m MatchResult) ParamsMap() VarsMap {
	if m.varsMap == nil {
		m.varsMap = make(VarsMap)
	}
	return m.varsMap
}

// GetValue returns the value of a named parameter and whether it was found.
// Returns the parameter value and true if the parameter exists, or empty string and false otherwise.
func (m MatchResult) GetValue(name common.Identifier) (value any, found bool) {
	value, found = m.varsMap[name]
	return value, found
}

// VarCount returns the number of extracted parameters.
func (m MatchResult) VarCount() int {
	return len(m.varsMap)
}

// HasVars returns true if any parameters were extracted from the request.
func (m MatchResult) HasVars() bool {
	return len(m.varsMap) > 0
}

// ForEachVar iterates over all extracted parameters, calling the provided function
// for each name-value pair. If the function returns true, iteration continues;
// if it returns false, iteration stops early.
func (m MatchResult) ForEachVar(fn func(name common.Identifier, value any) bool) {
	for name, value := range m.varsMap {
		if fn(name, value) {
			continue
		}
		goto end
	}
end:
	return
}
