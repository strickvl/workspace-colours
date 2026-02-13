package config

import (
	"fmt"
	"os/exec"
	"strings"
)

// verifyProcessCommand checks that a PID's command name contains the expected
// string. On macOS, we use `ps` since /proc doesn't exist.
func verifyProcessCommand(pid int, expectedCommand string) bool {
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=").Output()
	if err != nil {
		return false
	}
	comm := strings.TrimSpace(string(out))
	// Check if the command path contains the expected name.
	// e.g. comm="/Applications/Ghostty.app/.../ghostty", expectedCommand="ghostty"
	return strings.Contains(strings.ToLower(comm), strings.ToLower(expectedCommand))
}
