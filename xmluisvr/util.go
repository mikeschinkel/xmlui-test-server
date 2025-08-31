package xmluisvr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func launchBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // Unix-like
		cmd = "xdg-open"
		args = []string{url}
	}

	err := exec.Command(cmd, args...).Start()
	if err != nil {
		log.Printf("Failed to launch browser: %v", err)
	}
}

// Send error response with the given status code
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	log.Printf("Error: %s (Status: %d)", message, statusCode)
	http.Error(w, message, statusCode)
}

// Extract query parameters from request URL
func extractQueryParams(r *http.Request) map[string]string {
	queryParams := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}
	return queryParams
}

// Extract JSON body parameters from request
func extractBodyParams(r *http.Request) (map[string]any, error) {
	bodyParams := make(map[string]any)

	if r.Body == nil {
		return bodyParams, nil
	}

	var bodyBuffer bytes.Buffer
	bodyReader := io.TeeReader(r.Body, &bodyBuffer)

	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return bodyParams, err
	}

	if len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, &bodyParams)
		if err != nil {
			return bodyParams, err
		}
	}

	// Reset r.Body for potential future use
	r.Body = io.NopCloser(&bodyBuffer)

	return bodyParams, nil
}

// Convert a path template to a regexp
// Example: "/clients/:id" -> "^/clients/([^/]+)$"
func pathToRegexp(path string) string {
	// Escape any special regexp characters in the path
	escaped := regexp.QuoteMeta(path)

	// Replace :paramName with a capturing group
	re := regexp.MustCompile(`:([^/]+)`)
	regexpPath := re.ReplaceAllString(escaped, "([^/]+)")

	// Add start and end anchors
	return fmt.Sprintf("^%s$", regexpPath)
}

// Extract path parameters from a URL based on the endpoint path template
// Example: extractPathParams("/clients/123", "/clients/:id") -> {"id": "123"}
func extractPathParams(requestPath string, endpointPath string, re *regexp.Regexp) map[string]string {
	params := make(map[string]string)

	// Extract param names from the path template
	paramNames := make([]string, 0)
	pathParts := strings.Split(endpointPath, "/")
	for _, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			paramNames = append(paramNames, part[1:])
		}
	}

	// Extract values using regexp
	matches := re.FindStringSubmatch(requestPath)
	if len(matches) > 1 {
		// First match is the whole string, subsequent matches are capture groups
		for i, name := range paramNames {
			if i+1 < len(matches) {
				params[name] = matches[i+1]
			}
		}
	}

	return params
}

func closeOrLog(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("ERROR: Failed to close: %v", err)
	}
}

func nilOrLog(err error) {
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
}

func fprintf(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	nilOrLog(err)
}

func checkFileExists(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
		goto end
	}
	if info.IsDir() {
		err = errors.Join(ErrPathIsDir, err)
	}
end:
	return err
}
