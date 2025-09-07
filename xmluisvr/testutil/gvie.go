package testutil

// GVIE gets a value and ignores the error; used to just get value of func that
// returns T,error in a single expression. Use only in tests; never ignore errors
// otherwise,
func GVIE[T any](value T, _ error) T {
	return value
}
