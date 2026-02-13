package launcher

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/strickvl/workspace-colours/internal/color"
)

// ConfigureCursor writes color customizations into the project's
// .vscode/settings.json. It merges with any existing settings rather than
// overwriting the whole file.
func ConfigureCursor(scheme *color.Scheme, projectDir string) error {
	vsDir := filepath.Join(projectDir, ".vscode")
	if err := os.MkdirAll(vsDir, 0o755); err != nil {
		return fmt.Errorf("creating .vscode directory: %w", err)
	}

	settingsPath := filepath.Join(vsDir, "settings.json")

	// Load existing settings if present.
	settings := make(map[string]any)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			// If the file exists but is malformed, warn and start fresh.
			fmt.Fprintf(os.Stderr, "warning: could not parse %s, overwriting color settings\n", settingsPath)
		}
	}

	// Set the color customizations.
	settings["workbench.colorCustomizations"] = map[string]string{
		"titleBar.activeBackground":   scheme.Accent,
		"titleBar.activeForeground":   "#ffffff",
		"titleBar.inactiveBackground": scheme.AccentDim,
		"titleBar.inactiveForeground": "#cccccc",
		"activityBar.background":      scheme.AccentDim,
		"activityBar.foreground":      "#ffffff",
		"statusBar.background":        scheme.AccentDim,
		"statusBar.foreground":        "#ffffff",
		"sideBar.border":              scheme.Accent,
		"panel.border":                scheme.Accent,
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", settingsPath, err)
	}
	return nil
}

// LaunchCursor opens the project directory in Cursor.
func LaunchCursor(projectDir string) error {
	cursorBin, err := findCursor()
	if err != nil {
		return err
	}

	cmd := exec.Command(cursorBin, projectDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launching Cursor: %w", err)
	}
	return nil
}

// findCursor locates the Cursor CLI binary.
func findCursor() (string, error) {
	if path, err := exec.LookPath("cursor"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("cursor CLI not found in PATH — install it via Cursor → Command Palette → \"Install 'cursor' command in PATH\"")
}
