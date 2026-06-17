package browser

import (
	"os/exec"
	"runtime"
)

// Open launches the system default browser pointed at url. It returns as soon
// as the launcher process starts; it does not wait for the browser to exit.
// A non-nil error means the platform launcher could not be started (e.g. a
// headless box) — callers should fall back to printing the URL.
func Open(url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{url}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default: // linux, *bsd, etc.
		name, args = "xdg-open", []string{url}
	}
	return exec.Command(name, args...).Start()
}
