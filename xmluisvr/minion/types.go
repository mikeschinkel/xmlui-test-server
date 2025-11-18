package minion

import "github.com/mikeschinkel/go-dt"

// Source describes where to get the demo/project content
type Source struct {
	Type   string          // zip | release | clone | local
	Repo   string          // owner/repo format
	Branch string          // mutually exclusive with Tag
	Tag    string          // mutually exclusive with Branch
	Subdir dt.PathSegments // optional subdirectory within source
}

// CopyRule defines a file copy operation with glob support
type CopyRule struct {
	From     string // glob pattern
	To       string // destination path (can be file or dir)
	Optional bool
}

// Variant represents an alternative configuration
type Variant struct {
	Slug dt.PathSegment
	Name string
	Copy []CopyRule
}

// Registry represents the type-checked demo registry structure
type Registry struct {
	Schema      dt.URL
	Version     int
	DefaultSlug dt.PathSegment
	Demos       []Demo
}

// Demo represents a single demo entry in the registry
type Demo struct {
	Slug        dt.PathSegment
	Name        string
	ManifestURL dt.URL
}
