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

func fprintf(w io.Writer, format string, a ...any) int {
	n, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Printf("Failed to print to %v; %v", w, err)
	}
	return n
}
