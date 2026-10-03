package browser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Browser describes a single running Chromium-based browser process.
type Browser struct {
	Name        string
	PID         int
	ExePath     string
	UserDataDir string
	DebugPort   int
}

// DevToolsVersion is the JSON response of the DevTools /json/version endpoint.
type DevToolsVersion struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// browserBinNames lists the executable base names we treat as supported
// Chromium-based browsers.
var browserBinNames = []string{
	"chrome", "chrome.exe",
	"google-chrome", "google-chrome-stable",
	"chromium", "chromium-browser",
	"msedge", "msedge.exe", "microsoft-edge", "microsoft-edge.exe",
	"brave", "brave.exe",
}

// FindBrowsers returns every running Chromium-based browser process.
func FindBrowsers() ([]Browser, error) {
	return listBrowserProcesses()
}

// FindDevToolsEndpoint locates a usable DevTools HTTP endpoint. It first probes
// already-open debug ports, then (on POSIX) attempts to enable remote debugging
// on a running browser via SIGUSR1 and re-probes.
func FindDevToolsEndpoint() (string, error) {
	browsers, err := FindBrowsers()
	if err != nil {
		return "", err
	}
	for i := range browsers {
		if ep, e := browsers[i].DiscoverEndpoint(); e == nil {
			return ep, nil
		}
	}
	for i := range browsers {
		if e := EnableDebugging(&browsers[i]); e == nil {
			time.Sleep(600 * time.Millisecond)
			if ep, e := browsers[i].DiscoverEndpoint(); e == nil {
				return ep, nil
			}
		}
	}
	return "", fmt.Errorf("no browser connection: remote debugging is off; start the browser with --remote-debugging-port=9222")
}

// DiscoverEndpoint returns the DevTools HTTP base URL for this browser, or an
// error if no debug port is reachable.
func (b *Browser) DiscoverEndpoint() (string, error) {
	if ep := tryEndpoint("127.0.0.1:9222"); ep != "" {
		return ep, nil
	}
	if b.UserDataDir != "" {
		if port, ok := readDevToolsActivePort(b.UserDataDir); ok {
			if ep := tryEndpoint(fmt.Sprintf("127.0.0.1:%d", port)); ep != "" {
				return ep, nil
			}
		}
	}
	return "", fmt.Errorf("devtools endpoint yok: %s (pid %d)", b.Name, b.PID)
}

// Version fetches /json/version for the given endpoint.
func Version(endpoint string) (*DevToolsVersion, error) {
	c := &http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get(endpoint + "/json/version")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("devtools /json/version HTTP %d", resp.StatusCode)
	}
	var v DevToolsVersion
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func tryEndpoint(addr string) string {
	c := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := c.Get("http://" + addr + "/json/version")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	return "http://" + addr
}

func readDevToolsActivePort(userDataDir string) (int, bool) {
	data, err := os.ReadFile(filepath.Join(userDataDir, "DevToolsActivePort"))
	if err != nil {
		return 0, false
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return 0, false
	}
	port, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return 0, false
	}
	return port, true
}

func isBrowserBin(name string) bool {
	n := strings.ToLower(name)
	for _, b := range browserBinNames {
		if n == strings.ToLower(b) {
			return true
		}
	}
	return false
}

func parseUserDataSetDir(fields []string) string {
	for i, f := range fields {
		if f == "--user-data-dir" && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(f, "--user-data-dir=") {
			return strings.TrimPrefix(f, "--user-data-dir=")
		}
	}
	return ""
}
