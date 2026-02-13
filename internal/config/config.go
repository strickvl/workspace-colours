package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/strickvl/workspace-colours/internal/color"
)

const configDir = ".config/workspace-colours"
const assignmentsFile = "assignments.json"

// Assignment records which color scheme a project directory was given.
type Assignment struct {
	Scheme     string    `json:"scheme"`
	AssignedAt time.Time `json:"assigned_at"`
}

// Assignments maps absolute project paths to their color assignments.
type Assignments map[string]Assignment

// configPath returns the full path to the assignments file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, configDir, assignmentsFile), nil
}

// Load reads the assignments file from disk. Returns an empty map if the
// file doesn't exist yet.
func Load() (Assignments, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(Assignments), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var a Assignments
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return a, nil
}

// Save writes the assignments to disk, creating the directory if needed.
func Save(a Assignments) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding assignments: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// GetOrAssign looks up the color for a project. If none is assigned, it picks
// the next available color from the palette and persists the choice.
// If forceName is non-empty, it overrides any existing assignment.
func GetOrAssign(projectDir string, forceName string) (*color.Scheme, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	assignments, err := Load()
	if err != nil {
		return nil, err
	}

	// If a specific color was requested, use it.
	if forceName != "" {
		scheme := color.ByName(forceName)
		if scheme == nil {
			return nil, fmt.Errorf("unknown color scheme %q (available: %v)", forceName, color.Names())
		}
		assignments[absDir] = Assignment{Scheme: forceName, AssignedAt: time.Now()}
		if err := Save(assignments); err != nil {
			return nil, err
		}
		return scheme, nil
	}

	// Check for existing assignment.
	if a, ok := assignments[absDir]; ok {
		scheme := color.ByName(a.Scheme)
		if scheme != nil {
			return scheme, nil
		}
		// Scheme name in file no longer exists in palette — reassign.
	}

	// Find which colors are already in use.
	used := make(map[string]bool)
	for _, a := range assignments {
		used[a.Scheme] = true
	}

	// Pick the first unused color.
	for _, s := range color.Palettes {
		if !used[s.Name] {
			assignments[absDir] = Assignment{Scheme: s.Name, AssignedAt: time.Now()}
			if err := Save(assignments); err != nil {
				return nil, err
			}
			return color.ByName(s.Name), nil
		}
	}

	// All colors are in use — recycle from the beginning.
	scheme := &color.Palettes[0]
	assignments[absDir] = Assignment{Scheme: scheme.Name, AssignedAt: time.Now()}
	if err := Save(assignments); err != nil {
		return nil, err
	}
	return scheme, nil
}

// Reset removes the color assignment for a project.
func Reset(projectDir string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	assignments, err := Load()
	if err != nil {
		return err
	}

	delete(assignments, absDir)
	return Save(assignments)
}

// List returns all current assignments.
func List() (Assignments, error) {
	return Load()
}
