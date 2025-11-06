package xmluisvr

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	. "github.com/mikeschinkel/go-doterr"
)

// launchBrowser attempts to open a URL in the default web browser.
// It uses platform-specific commands to launch the browser:
//   - macOS: 'open'
//   - Windows: 'rundll32'
//   - Unix-like: 'xdg-open'
//
// If the browser fails to launch, an error is logged but not returned.
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

// checkFileExists verifies if a file exists at the given path and whether it's a directory.
// Returns ErrPathIsDir if the path exists but is a directory.
// Returns nil if the file doesn't exist (this is not considered an error).
func checkFileExists(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
		goto end
	}
	if info.IsDir() {
		err = NewErr(ErrPathIsDir, err)
	}
end:
	return err
}

// fprintf is a wrapper around fmt.Fprintf that logs errors instead of returning them.
// This is used for non-critical output operations where error handling would clutter the code.
// Returns the number of bytes written.
func fprintf(w io.Writer, format string, a ...any) int {
	n, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Printf("Failed to print to %v; %v", w, err)
	}
	return n
}
