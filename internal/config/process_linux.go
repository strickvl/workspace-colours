package config

import (
	"os"
	"fmt"
	"strings"
)

// verifyProcessCommand checks that a PID's command name contains the expected
// string. On Linux, we read /proc/<pid>/comm.
func verifyProcessCommand(pid int, expectedCommand string) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return false
	}
	comm := strings.TrimSpace(string(data))
	return strings.Contains(strings.ToLower(comm), strings.ToLower(expectedCommand))
}
