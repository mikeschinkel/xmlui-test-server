package minion

import (
	"fmt"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtglob"
	"github.com/xmlui-org/localsvr/xmluisvr/cfgldr"
)

// ParseManifest converts raw cfgldr.Manifest to type-checked runpkg.Manifest
func ParseManifest(raw *cfgldr.Manifest) (manifest *Manifest, err error) {
	var schema dt.URL
	var slug dt.PathSegment
	var subdir dt.PathSegments
	var copyRules []CopyRule
	var variants []Variant

	// Parse Schema URL
	schema, err = dt.ParseURL(raw.Schema)
	if err != nil {
		err = fmt.Errorf("invalid schema URL: %w", err)
		goto end
	}

	// Parse Slug
	slug, err = dt.ParsePathSegment(raw.Slug)
	if err != nil {
		err = fmt.Errorf("invalid slug: %w", err)
		goto end
	}

	// Parse Subdir if present
	if raw.Source.Subdir != "" {
		subdir = dt.PathSegments(raw.Source.Subdir)
	}

	// Convert copy copyRules
	copyRules, err = parseCopyRules(raw.Copy)
	if err != nil {
		goto end
	}

	// Convert variants
	variants, err = parseVariants(raw.Variants)
	if err != nil {
		goto end
	}

	manifest = &Manifest{
		Schema:      schema,
		Version:     raw.Version,
		Slug:        slug,
		Name:        raw.Name,
		Description: raw.Description,
		Source: Source{
			Type:   raw.Source.Type,
			Repo:   raw.Source.Repo,
			Branch: raw.Source.Branch,
			Tag:    raw.Source.Tag,
			Subdir: subdir,
		},
		Copy:     copyRules,
		Variants: variants,
	}

	// Validate manifest
	err = validateManifest(manifest)
	if err != nil {
		goto end
	}

end:
	return manifest, err
}

// ParseRegistry converts raw cfgldr.Registry to type-checked runpkg.Registry
func ParseRegistry(raw *cfgldr.DemoRegistry) (registry *Registry, err error) {
	var schema dt.URL
	var defaultSlug dt.PathSegment
	var demos []Demo

	// Parse Schema URL
	schema, err = dt.ParseURL(raw.Schema)
	if err != nil {
		err = fmt.Errorf("invalid schema URL: %w", err)
		goto end
	}

	// Parse DefaultSlug
	defaultSlug, err = dt.ParsePathSegment(raw.DefaultSlug)
	if err != nil {
		err = fmt.Errorf("invalid default_slug: %w", err)
		goto end
	}

	// Parse Demos
	demos, err = parseDemos(raw.Demos)
	if err != nil {
		goto end
	}

	registry = &Registry{
		Schema:      schema,
		Version:     raw.Version,
		DefaultSlug: defaultSlug,
		Demos:       demos,
	}

end:
	return registry, err
}

// parseCopyRules converts raw copy rules to type-checked copy rules
func parseCopyRules(rawRules []cfgldr.CopyRule) (rules []CopyRule, err error) {
	var rawRule cfgldr.CopyRule
	var i int

	rules = make([]CopyRule, len(rawRules))
	for i, rawRule = range rawRules {
		rules[i] = CopyRule{
			From:     rawRule.From,
			To:       rawRule.To,
			Optional: rawRule.Optional,
		}
	}

	return rules, err
}

// parseVariants converts raw variants to type-checked variants
func parseVariants(rawVariants []cfgldr.Variant) (variants []Variant, err error) {
	var rawVariant cfgldr.Variant
	var slug dt.PathSegment
	var copyRules []CopyRule
	var i int

	variants = make([]Variant, len(rawVariants))
	for i, rawVariant = range rawVariants {
		slug, err = dt.ParsePathSegment(rawVariant.Slug)
		if err != nil {
			err = fmt.Errorf("invalid variant slug: %w", err)
			goto end
		}

		copyRules, err = parseCopyRules(rawVariant.Copy)
		if err != nil {
			goto end
		}

		variants[i] = Variant{
			Slug: slug,
			Name: rawVariant.Name,
			Copy: copyRules,
		}
	}

end:
	return variants, err
}

// parseDemos converts raw demos to type-checked demos
func parseDemos(rawDemos []cfgldr.Demo) (demos []Demo, err error) {
	var rawDemo cfgldr.Demo
	var slug dt.PathSegment
	var manifestURL dt.URL
	var i int

	demos = make([]Demo, len(rawDemos))
	for i, rawDemo = range rawDemos {
		slug, err = dt.ParsePathSegment(rawDemo.Slug)
		if err != nil {
			err = fmt.Errorf("invalid demo slug: %w", err)
			goto end
		}

		manifestURL, err = dt.ParseURL(rawDemo.ManifestURL)
		if err != nil {
			err = fmt.Errorf("invalid manifest URL: %w", err)
			goto end
		}

		demos[i] = Demo{
			Slug:        slug,
			Name:        rawDemo.Name,
			ManifestURL: manifestURL,
		}
	}

end:
	return demos, err
}

// validateManifest performs basic validation on the manifest structure
func validateManifest(m *Manifest) (err error) {
	var validTypes map[string]bool

	if m.Slug == "" {
		err = fmt.Errorf("manifest missing required field: slug")
		goto end
	}
	if m.Name == "" {
		err = fmt.Errorf("manifest missing required field: name")
		goto end
	}
	if m.Source.Type == "" {
		err = fmt.Errorf("manifest missing required field: source.type")
		goto end
	}
	if m.Source.Repo == "" {
		err = fmt.Errorf("manifest missing required field: source.repo")
		goto end
	}

	// Validate source type (v0 only supports zip and release)
	validTypes = map[string]bool{
		"zip":     true,
		"release": true,
	}
	if !validTypes[m.Source.Type] {
		err = fmt.Errorf("unsupported source type '%s' (v0 supports: zip, release)", m.Source.Type)
		goto end
	}

	// Validate that branch OR tag is specified, not both
	if m.Source.Branch != "" && m.Source.Tag != "" {
		err = fmt.Errorf("source cannot specify both branch and tag")
		goto end
	}

	// At least one copy rule is recommended
	if len(m.Copy) == 0 {
		err = fmt.Errorf("manifest has no copy rules")
		goto end
	}

end:
	return err
}

// ParseGlobRules converts application CopyRule to dtglob.GlobRules
func ParseGlobRules(rules []CopyRule, baseDir dt.DirPath) (globRules *dtglob.GlobRules, err error) {
	var dtRules []dtglob.GlobRule
	var i int
	var r CopyRule

	dtRules = make([]dtglob.GlobRule, len(rules))
	for i, r = range rules {
		dtRules[i] = dtglob.GlobRule{
			From:     dtglob.Glob(r.From),
			To:       dt.EntryPath(r.To),
			Optional: r.Optional,
		}
	}

	globRules = &dtglob.GlobRules{
		BaseDir: baseDir,
		Rules:   dtRules,
	}

	return globRules, err
}
