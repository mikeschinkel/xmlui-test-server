package minion

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mikeschinkel/go-dt"
	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cfgldr"
)

// InputType represents the kind of input provided by the user
type InputType int

const (
	InputTypeUnspecified InputType = iota
	InputTypeName                  // e.g., "xmlui-invoice"
	InputTypeManifestURL           // e.g., "https://raw.githubusercontent.com/.../xmlui-manifest.json"
	InputTypeRepoURL               // e.g., "https://github.com/xmlui-org/xmlui-invoice"
	InputTypeOwnerRepo             // e.g., "xmlui-org/xmlui-invoice"
)

// ResolveManifestURL takes user input and resolves it to a manifest URL
func ResolveManifestURL(input string, configDir dt.DirPath) (u dt.URL, err error) {
	var rawRegistry *cfgldr.DemoRegistry
	var registry *Registry
	var inputType InputType
	var ownerRepo string

	if input == "" {
		// No input - use default from registry
		rawRegistry, err = cfgldr.LoadDemoRegistry(string(configDir))
		if err != nil {
			err = fmt.Errorf("failed to load registry: %w", err)
			goto end
		}

		registry, err = ParseRegistry(rawRegistry)
		if err != nil {
			err = fmt.Errorf("failed to parse registry: %w", err)
			goto end
		}

		u, err = registry.GetDefault()
		if err != nil {
			err = fmt.Errorf("failed to get default demo: %w", err)
			goto end
		}
		goto end
	}

	inputType = detectInputType(input)

	switch inputType {
	case InputTypeName:
		// Lookup in registry
		rawRegistry, err = cfgldr.LoadDemoRegistry(string(configDir))
		if err != nil {
			err = fmt.Errorf("failed to load registry: %w", err)
			goto end
		}

		registry, err = ParseRegistry(rawRegistry)
		if err != nil {
			err = fmt.Errorf("failed to parse registry: %w", err)
			goto end
		}

		u, err = registry.Lookup(dt.PathSegment(input))
		if err != nil {
			err = fmt.Errorf("failed to lookup demo '%s': %w", input, err)
			goto end
		}

	case InputTypeManifestURL:
		// Direct manifest URL - use as is
		u, err = dt.ParseURL(input)
		if err != nil {
			err = fmt.Errorf("invalid manifest URL: %w", err)
			goto end
		}

	case InputTypeRepoURL:
		// Extract owner/repo from URL
		ownerRepo, err = extractOwnerRepo(input)
		if err != nil {
			goto end
		}
		u, err = resolveFromRepo(ownerRepo)
		if err != nil {
			goto end
		}

	case InputTypeOwnerRepo:
		// Apply repo convention
		u, err = resolveFromRepo(input)
		if err != nil {
			goto end
		}

	default:
		err = fmt.Errorf("unable to resolve input: %s", input)
		goto end
	}

end:
	return u, err
}

// GetDefault returns the default demo's manifest URL
func (r *Registry) GetDefault() (u dt.URL, err error) {
	if r.DefaultSlug == "" {
		err = fmt.Errorf("no default demo configured in registry")
		goto end
	}
	u, err = r.Lookup(r.DefaultSlug)

end:
	return u, err
}

// Lookup finds a demo by slug and returns its manifest URL
func (r *Registry) Lookup(slug dt.PathSegment) (u dt.URL, err error) {
	var demo Demo

	for _, demo = range r.Demos {
		if demo.Slug == slug {
			u = dt.URL(demo.ManifestURL)
			goto end
		}
	}
	err = fmt.Errorf("demo '%s' not found in registry", slug)

end:
	return u, err
}

// detectInputType determines what kind of input was provided
func detectInputType(input string) InputType {
	// Check for URL patterns
	if //goland:noinspection HttpUrlsUsage
	strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		// Check if it's a manifest URL
		if strings.HasSuffix(input, "xmlui-manifest.json") {
			return InputTypeManifestURL
		}
		// Otherwise it's a repo URL
		return InputTypeRepoURL
	}

	// Check for owner/repo format (contains exactly one /)
	if strings.Contains(input, "/") && strings.Count(input, "/") == 1 {
		return InputTypeOwnerRepo
	}

	// Simple slug name (no slashes, no URLs)
	if !strings.Contains(input, "/") && !strings.Contains(input, ".") {
		return InputTypeName
	}

	return InputTypeUnspecified
}

// extractOwnerRepo extracts the owner/repo from a GitHub URL
func extractOwnerRepo(url string) (ownerRepo string, err error) {
	var parts []string

	// Handle github.com URLs
	// Examples:
	// https://github.com/xmlui-org/xmlui-invoice
	// https://github.com/xmlui-org/xmlui-invoice/tree/branch
	url = strings.TrimPrefix(url, "https://github.com/")
	//goland:noinspection HttpUrlsUsage
	url = strings.TrimPrefix(url, "http://github.com/")

	parts = strings.Split(url, "/")
	if len(parts) < 2 {
		err = fmt.Errorf("invalid GitHub URL: missing owner/repo")
		goto end
	}

	ownerRepo = fmt.Sprintf("%s/%s", parts[0], parts[1])

end:
	return ownerRepo, err
}

// resolveFromRepo applies the repo convention to find manifest
// Tries in order: release branch, main branch
func resolveFromRepo(ownerRepo string) (u dt.URL, err error) {
	var refs []string
	var ref string
	var manifestURL string

	refs = []string{"release", "main"}

	for _, ref = range refs {
		manifestURL = fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/xmlui-manifest.json",
			ownerRepo,
			ref,
		)

		// Try to fetch the manifest
		if manifestExists(manifestURL) {
			u, err = dt.ParseURL(manifestURL)
			if err != nil {
				err = fmt.Errorf("failed to parse manifest URL: %w", err)
				goto end
			}
			goto end
		}
	}

	err = fmt.Errorf(
		"manifest not found in repo %s (tried refs: %s). "+
			"Use --manifest flag to specify a direct manifest URL, or --branch/--tag to specify a different ref",
		ownerRepo,
		strings.Join(refs, ", "),
	)

end:
	return u, err
}

// manifestExists checks if a manifest URL returns 200 OK
func manifestExists(url string) bool {
	var resp *http.Response
	var err error

	resp, err = http.Head(url)
	if err != nil {
		return false
	}
	defer dt.CloseOrLog(resp.Body)
	return resp.StatusCode == http.StatusOK
}
