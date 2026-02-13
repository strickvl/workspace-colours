package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/strickvl/workspace-colours/internal/color"
)

const firefoxProfilePrefix = "workspace-"

// FirefoxVariant represents which Firefox binary to use.
type FirefoxVariant int

const (
	FirefoxStandard FirefoxVariant = iota
	FirefoxDeveloperEdition
)

// firefoxBinaryPaths returns candidate paths for the Firefox binary.
func firefoxBinaryPaths(variant FirefoxVariant) []string {
	if runtime.GOOS != "darwin" {
		// Linux paths (best-effort).
		return []string{"firefox", "firefox-developer-edition"}
	}
	switch variant {
	case FirefoxDeveloperEdition:
		return []string{
			"/Applications/Firefox Developer Edition.app/Contents/MacOS/firefox",
		}
	default:
		return []string{
			"/Applications/Firefox.app/Contents/MacOS/firefox-bin",
			"/Applications/Firefox.app/Contents/MacOS/firefox",
		}
	}
}

// findFirefox locates the Firefox binary, preferring Developer Edition if
// available, falling back to standard Firefox.
func findFirefox() (string, FirefoxVariant, error) {
	// Try Developer Edition first.
	for _, p := range firefoxBinaryPaths(FirefoxDeveloperEdition) {
		if _, err := os.Stat(p); err == nil {
			return p, FirefoxDeveloperEdition, nil
		}
	}
	// Fall back to standard.
	for _, p := range firefoxBinaryPaths(FirefoxStandard) {
		if _, err := os.Stat(p); err == nil {
			return p, FirefoxStandard, nil
		}
	}
	// Check PATH.
	if path, err := exec.LookPath("firefox"); err == nil {
		return path, FirefoxStandard, nil
	}
	return "", FirefoxStandard, fmt.Errorf("Firefox not found — install Firefox or Firefox Developer Edition")
}

// firefoxProfilesDir returns the base directory where Firefox stores profiles.
// Both standard Firefox and Developer Edition share the same profiles directory.
func firefoxProfilesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles"), nil
	}
	return filepath.Join(home, ".mozilla", "firefox"), nil
}

// profileName returns the Firefox profile name for a colour scheme.
func profileName(scheme *color.Scheme) string {
	return firefoxProfilePrefix + scheme.Name
}

// findProfileDir searches for an existing Firefox profile directory matching
// the scheme. Profile directories have a random prefix: <random>.<profileName>.
func findProfileDir(scheme *color.Scheme) (string, error) {
	baseDir, err := firefoxProfilesDir()
	if err != nil {
		return "", err
	}

	name := profileName(scheme)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("reading profiles directory %s: %w", baseDir, err)
	}

	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), "."+name) {
			return filepath.Join(baseDir, e.Name()), nil
		}
	}
	return "", nil
}

// findProfileDirFromINI looks up the profile path from profiles.ini as a
// fallback when the directory naming convention doesn't match.
func findProfileDirFromINI(scheme *color.Scheme) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var iniPath string
	if runtime.GOOS == "darwin" {
		iniPath = filepath.Join(home, "Library", "Application Support", "Firefox", "profiles.ini")
	} else {
		iniPath = filepath.Join(home, ".mozilla", "firefox", "profiles.ini")
	}

	f, err := os.Open(iniPath)
	if err != nil {
		return "", fmt.Errorf("opening profiles.ini: %w", err)
	}
	defer f.Close()

	name := profileName(scheme)
	scanner := bufio.NewScanner(f)
	var currentPath string
	var isRelative bool
	var foundName bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			// New section — check if previous section matched.
			if foundName && currentPath != "" {
				if isRelative {
					return filepath.Join(filepath.Dir(iniPath), currentPath), nil
				}
				return currentPath, nil
			}
			currentPath = ""
			isRelative = false
			foundName = false
			continue
		}
		if strings.EqualFold(line, "Name="+name) {
			foundName = true
		}
		if strings.HasPrefix(line, "Path=") {
			currentPath = strings.TrimPrefix(line, "Path=")
		}
		if line == "IsRelative=1" {
			isRelative = true
		}
	}

	// Check last section.
	if foundName && currentPath != "" {
		if isRelative {
			return filepath.Join(filepath.Dir(iniPath), currentPath), nil
		}
		return currentPath, nil
	}

	return "", nil
}

// EnsureFirefoxProfile creates a Firefox profile for the colour scheme if it
// doesn't exist, and writes the userChrome.css theming.
func EnsureFirefoxProfile(scheme *color.Scheme) error {
	bin, _, err := findFirefox()
	if err != nil {
		return err
	}

	name := profileName(scheme)

	// Check if profile already exists.
	profileDir, err := findProfileDir(scheme)
	if err != nil {
		return err
	}

	if profileDir == "" {
		// Try profiles.ini as fallback.
		profileDir, _ = findProfileDirFromINI(scheme)
	}

	if profileDir == "" {
		// Create the profile.
		fmt.Printf("Creating Firefox profile %q...\n", name)
		cmd := exec.Command(bin, "-CreateProfile", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("creating Firefox profile: %w", err)
		}

		// Find the newly created directory.
		profileDir, err = findProfileDir(scheme)
		if err != nil || profileDir == "" {
			profileDir, _ = findProfileDirFromINI(scheme)
		}
		if profileDir == "" {
			return fmt.Errorf("could not find profile directory after creating profile %q", name)
		}
	}

	// Write userChrome.css.
	chromeDir := filepath.Join(profileDir, "chrome")
	if err := os.MkdirAll(chromeDir, 0o755); err != nil {
		return fmt.Errorf("creating chrome directory: %w", err)
	}

	css := fmt.Sprintf(`:root {
  --toolbar-bgcolor: %s !important;
  --toolbar-color: %s !important;
  --tab-selected-bgcolor: %s !important;
  --tab-selected-textcolor: #ffffff !important;
  --lwt-accent-color: #%s !important;
  --lwt-text-color: %s !important;
}

#TabsToolbar {
  background-color: #%s !important;
}

#nav-bar {
  background-color: %s !important;
}

.tabbrowser-tab[selected="true"] {
  background-color: %s !important;
}
`, scheme.Accent, "#"+scheme.GhosttyFG, scheme.Accent,
		scheme.GhosttyBG, "#"+scheme.GhosttyFG,
		scheme.GhosttyBG, scheme.AccentDim, scheme.Accent)

	cssPath := filepath.Join(chromeDir, "userChrome.css")
	if err := os.WriteFile(cssPath, []byte(css), 0o644); err != nil {
		return fmt.Errorf("writing userChrome.css: %w", err)
	}

	// Enable userChrome.css via user.js preference.
	userJS := `user_pref("toolkit.legacyUserProfileCustomizations.stylesheets", true);
`
	userJSPath := filepath.Join(profileDir, "user.js")
	if err := os.WriteFile(userJSPath, []byte(userJS), 0o644); err != nil {
		return fmt.Errorf("writing user.js: %w", err)
	}

	return nil
}

// LaunchFirefox opens Firefox with the workspace-themed profile.
func LaunchFirefox(scheme *color.Scheme) error {
	bin, _, err := findFirefox()
	if err != nil {
		return err
	}

	name := profileName(scheme)
	cmd := exec.Command(bin, "-P", name, "-no-remote")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launching Firefox with profile %q: %w", name, err)
	}
	return nil
}
