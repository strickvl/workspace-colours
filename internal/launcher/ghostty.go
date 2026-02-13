package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/strickvl/workspace-colours/internal/color"
)

const ghosttyThemeDir = ".config/ghostty/themes"
const ghosttyThemePrefix = "workspace-"

// themeFileName returns the Ghostty theme file name for a color scheme.
func themeFileName(scheme *color.Scheme) string {
	return ghosttyThemePrefix + scheme.Name
}

// EnsureGhosttyTheme writes a Ghostty theme file for the given scheme if it
// doesn't already exist. Theme files live in ~/.config/ghostty/themes/.
func EnsureGhosttyTheme(scheme *color.Scheme) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}

	dir := filepath.Join(home, ghosttyThemeDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating theme directory: %w", err)
	}

	path := filepath.Join(dir, themeFileName(scheme))

	content := fmt.Sprintf(`background = %s
foreground = %s
cursor-color = %s
selection-background = %s
selection-foreground = ffffff
`, scheme.GhosttyBG, scheme.GhosttyFG, scheme.CursorColor, scheme.SelectionBG)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing theme %s: %w", path, err)
	}
	return nil
}

// LaunchGhostty opens N Ghostty terminal windows with the given color scheme.
// Each window gets a title derived from the project name and an index.
func LaunchGhostty(scheme *color.Scheme, projectDir string, count int) error {
	if err := EnsureGhosttyTheme(scheme); err != nil {
		return err
	}

	ghosttyBin, err := findGhostty()
	if err != nil {
		return err
	}

	projectName := filepath.Base(projectDir)
	themeName := themeFileName(scheme)

	labels := defaultTerminalLabels(count)
	for i := range count {
		title := fmt.Sprintf("%s — %s", projectName, labels[i])
		args := []string{
			fmt.Sprintf("--theme=%s", themeName),
			fmt.Sprintf("--title=%s", title),
			fmt.Sprintf("--working-directory=%s", projectDir),
		}

		cmd := exec.Command(ghosttyBin, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("launching Ghostty window %d: %w", i+1, err)
		}
		// Don't wait — these are background GUI processes.
	}
	return nil
}

// findGhostty locates the Ghostty binary.
func findGhostty() (string, error) {
	// Check PATH first.
	if path, err := exec.LookPath("ghostty"); err == nil {
		return path, nil
	}
	// Common macOS location.
	macPath := "/Applications/Ghostty.app/Contents/MacOS/ghostty"
	if _, err := os.Stat(macPath); err == nil {
		return macPath, nil
	}
	return "", fmt.Errorf("ghostty not found in PATH or /Applications — is it installed?")
}

// defaultTerminalLabels returns human-friendly labels for N terminal windows.
func defaultTerminalLabels(n int) []string {
	defaults := []string{"Main", "Server", "Tests", "Git"}
	if n <= len(defaults) {
		return defaults[:n]
	}
	labels := make([]string, n)
	copy(labels, defaults)
	for i := len(defaults); i < n; i++ {
		labels[i] = fmt.Sprintf("Terminal %d", i+1)
	}
	return labels
}

// ListInstalledThemes returns workspace theme names currently installed.
func ListInstalledThemes() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ghosttyThemeDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ghosttyThemePrefix) {
			themes = append(themes, e.Name())
		}
	}
	return themes, nil
}
