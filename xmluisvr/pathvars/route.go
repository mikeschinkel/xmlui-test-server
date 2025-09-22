package pathvars

// Route represents a compiled endpoint
type Route struct {
	Method   string
	Template *Template
	Index    int
}
