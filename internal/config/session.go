package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const sessionsDir = "sessions"

// TrackedProcess records a launched process so we can find it later.
type TrackedProcess struct {
	PID         int       `json:"pid"`
	CommandName string    `json:"command_name"` // e.g. "ghostty", "firefox", "cursor"
	Description string    `json:"description"`  // human label, e.g. "Ghostty — Main"
	LaunchedAt  time.Time `json:"launched_at"`
}

// Session records all processes launched for a workspace.
type Session struct {
	ProjectDir string            `json:"project_dir"`
	Scheme     string            `json:"scheme"`
	Processes  []TrackedProcess  `json:"processes"`
	CreatedAt  time.Time         `json:"created_at"`
}

// sessionPath returns the file path for a project's session file.
// Uses a SHA256 hash of the absolute path to avoid filesystem issues.
func sessionPath(projectDir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(projectDir)))[:16]
	return filepath.Join(home, configDir, sessionsDir, hash+".json"), nil
}

// LoadSession reads the session file for a project. Returns nil if no session exists.
func LoadSession(projectDir string) (*Session, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	path, err := sessionPath(absDir)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading session %s: %w", path, err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}
	return &s, nil
}

// SaveSession writes the session to disk.
func SaveSession(s *Session) error {
	path, err := sessionPath(s.ProjectDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating sessions directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// DeleteSession removes the session file for a project.
func DeleteSession(projectDir string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return err
	}

	path, err := sessionPath(absDir)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ListSessions returns all active sessions.
func ListSessions() ([]*Session, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, configDir, sessionsDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// IsProcessAlive checks if a PID is still running AND matches the expected
// command name. This prevents us from killing an unrelated process that
// inherited the same PID.
func IsProcessAlive(p TrackedProcess) bool {
	// First check: is the process alive at all?
	proc, err := os.FindProcess(p.PID)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. Send signal 0 to check existence.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	// Second check: does the command name match?
	// Read the process command from /proc or use ps on macOS.
	return verifyProcessCommand(p.PID, p.CommandName)
}

// KillProcess sends SIGTERM to a tracked process after verifying it's still
// the process we launched.
func KillProcess(p TrackedProcess) error {
	if !IsProcessAlive(p) {
		return nil // Already gone — nothing to do.
	}

	proc, err := os.FindProcess(p.PID)
	if err != nil {
		return nil
	}
	return proc.Signal(syscall.SIGTERM)
}
