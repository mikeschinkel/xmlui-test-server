package pathvars

type Selector string
type Identifier string
type Location string
type HTTPMethod string

func Identifiers[S ~string](ss []S) (ids []Identifier) {
	ids = make([]Identifier, len(ss))
	for i, s := range ss {
		ids[i] = Identifier(s)
	}
	return ids
}
