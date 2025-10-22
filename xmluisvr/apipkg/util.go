package apipkg

// combineStringsAsX combines two string slices that both derive from string and
// return the combined form as type []X.
func combineStringsAsX[X ~string, Y ~string](xx []X, yy []Y) []X {
	yyAsX := make([]X, len(yy))
	for i, y := range yy {
		yyAsX[i] = X(y)
	}
	return append(xx, yyAsX...)
}

// combineStringsAsY combines two string slices that both derive from string and
// return the combined form as type []Y.
func combineStringsAsY[X ~string, Y ~string](xx []X, yy []Y) []Y {
	xxAsY := make([]Y, len(xx))
	for i, x := range xx {
		xxAsY[i] = Y(x)
	}
	return append(yy, xxAsY...)
}

func prepend[T any](slice []T, item ...T) []T {
	return append(item, slice...)
}
