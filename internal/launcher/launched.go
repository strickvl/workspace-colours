package launcher

// LaunchedProcess records a process that was started by a launcher,
// so the session tracker can store it and close it later.
type LaunchedProcess struct {
	PID         int
	CommandName string // short name, e.g. "ghostty", "firefox", "cursor"
	Description string // human label, e.g. "Ghostty — Main"
}
