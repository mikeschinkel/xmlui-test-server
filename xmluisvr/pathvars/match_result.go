// Package pathvars/match_result defines the MatchResult type which contains
// the results of matching an HTTP request against a route template.
// It provides access to extracted parameter values and route information.
package pathvars

import (
	"maps"
	"slices"
)

// ValuesMap is an ordered map of parameter names to their extracted values.
// Order preservation is critical for:
//   - Error suggestion URLs that match the user's request parameter order (ADR-018)
//   - Deterministic test behavior (no map iteration randomness)
//   - Debug output that reflects actual HTTP request structure
type ValuesMap struct {
	*OrderedMap[Identifier, any]
}

func (vm ValuesMap) SetNil() {
	vm.OrderedMap = nil
}

func (vm ValuesMap) IsNil() bool {
	return vm.OrderedMap == nil
}

func NewValuesMap(cap int) ValuesMap {
	return ValuesMap{
		OrderedMap: NewOrderedMap[Identifier, any](cap),
	}
}

func (vm ValuesMap) GetValues(names []Identifier) (values ValuesMap, notFound []Identifier) {
	n := len(names)
	values = NewValuesMap(n)
	notFound = make([]Identifier, 0, n)

	notFoundMap := make(map[Identifier]struct{}, len(names))
	for _, name := range names {
		notFoundMap[name] = struct{}{}
	}
	for _, name := range names {
		value, ok := vm.Get(name)
		if !ok {
			continue
		}
		values.Set(name, value)
		delete(notFoundMap, name)
	}
	notFound = slices.Collect(maps.Keys(notFoundMap))
	return values, notFound
}

// MatchResult represents the result of matching an HTTP request against a route template.
// It contains the matched route index and extracted parameter values for memory efficiency.
type MatchResult struct {
	// Index indicates which route was matched in the router's route list.
	Index int

	Route *Route

	// valuesMap contains the extracted parameter values from the matched request.
	// This field is private to control access and ensure proper initialization.
	valuesMap ValuesMap
}

// NewMatchResult creates a new MatchResult with the specified route index and parameter values.
func NewMatchResult(r *Route, valuesMap ValuesMap) MatchResult {
	return MatchResult{
		Index:     r.Index,
		Route:     r,
		valuesMap: valuesMap,
	}
}

func (m MatchResult) GetValues(names []Identifier) (ValuesMap, []Identifier) {
	return m.valuesMap.GetValues(names)
}

// ValuesMap returns the map of extracted parameter values.
// If the internal map is nil, it initializes an empty map to prevent nil pointer issues.
func (m MatchResult) ValuesMap() ValuesMap {
	if m.valuesMap.IsNil() {
		m.valuesMap = NewValuesMap(0)
	}
	return m.valuesMap
}

// GetValue returns the value of a named parameter and whether it was found.
// Returns the parameter value and true if the parameter exists, or empty string and false otherwise.
func (m MatchResult) GetValue(name Identifier) (value any, found bool) {
	value, found = m.valuesMap.Get(name)
	return value, found
}

// VarCount returns the number of extracted parameters.
func (m MatchResult) VarCount() int {
	return m.valuesMap.Len()
}

// HasVars returns true if any parameters were extracted from the request.
func (m MatchResult) HasVars() bool {
	return m.valuesMap.Len() > 0
}

// ForEachVar iterates over all extracted parameters, calling the provided function
// for each name-value pair. If the function returns true, iteration continues;
// if it returns false, iteration stops early.
func (m MatchResult) ForEachVar(fn func(name Identifier, value any) bool) {
	for name, value := range m.valuesMap.Iterator() {
		if fn(name, value) {
			continue
		}
		goto end
	}
end:
	return
}
