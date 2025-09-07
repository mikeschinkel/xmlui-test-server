package xmluisvr

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
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
