package minion

import (
	"github.com/mikeschinkel/go-dt"
)

// Manifest represents the type-checked xmlui-manifest.json structure
type Manifest struct {
	Schema      dt.URL
	Version     int
	Slug        dt.PathSegment
	Name        string
	Description string
	Source      Source
	Copy        []CopyRule
	Variants    []Variant
}
