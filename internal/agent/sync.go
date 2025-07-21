package agent

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// ShowDashboard opens the ContextKeeper.dev dashboard with the current session
func (a *Agent) ShowDashboard() error {
	sessionID := a.sessionMgr.GetOrCreateSession()
	url := fmt.Sprintf("https://contextkeeper.dev/app?session=%s", sessionID)
	
	log.Printf("Opening ContextKeeper dashboard: %s", url)
	return openBrowser(url)
}

// GetSessionURL returns the dashboard URL with session ID
func (a *Agent) GetSessionURL() string {
	sessionID := a.sessionMgr.GetOrCreateSession()
	return fmt.Sprintf("https://contextkeeper.dev/app?session=%s", sessionID)
}

// ShowSessionInfo displays session information to the user
func (a *Agent) ShowSessionInfo() {
	sessionID := a.sessionMgr.GetOrCreateSession()
	url := a.GetSessionURL()
	
	fmt.Printf("🔗 ContextKeeper Session\n")
	fmt.Printf("========================\n")
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Printf("Dashboard:  %s\n", url)
	fmt.Printf("\nTo view your AI sessions, visit the dashboard URL above.\n")
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}