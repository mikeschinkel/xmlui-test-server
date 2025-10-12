package jsonxtractr

type Selectors []Selector

func (ss Selectors) Strings() (strings []string) {
	strings = make([]string, len(ss))
	for i, s := range ss {
		strings[i] = string(s)
	}
	return strings
}

type Selector string

func ToSelectors[S ~string](ss []S) (ids []Selector) {
	ids = make([]Selector, len(ss))
	for i, s := range ss {
		ids[i] = Selector(s)
	}
	return ids
}
