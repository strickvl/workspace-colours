package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/strickvl/workspace-colours/internal/color"
	"github.com/strickvl/workspace-colours/internal/config"
	"github.com/strickvl/workspace-colours/internal/launcher"
)

func main() {
	terminals := flag.IntP("terminals", "t", 2, "number of Ghostty terminal windows to open")
	colorName := flag.StringP("color", "c", "", "force a specific color scheme (e.g. red, blue, green)")
	browser := flag.BoolP("browser", "b", false, "also launch a color-themed Firefox profile")
	list := flag.BoolP("list", "l", false, "list all current color assignments")
	resetColor := flag.Bool("reset-color", false, "remove the color assignment for a project")
	noCursor := flag.Bool("no-cursor", false, "skip opening Cursor IDE")
	noTerminals := flag.Bool("no-terminals", false, "skip opening Ghostty terminals")
	borders := flag.Bool("borders", false, "update JankyBorders active window colour")
	closeAll := flag.Bool("close-all", false, "close all tracked workspace windows")
	flag.Usage = usage

	flag.Parse()

	if *list {
		runList()
		return
	}

	if *closeAll {
		runCloseAll()
		return
	}

	// Check for subcommands.
	if flag.Arg(0) == "close" {
		projectDir := flag.Arg(1)
		if projectDir == "" {
			fatalf("usage: workspace close <project-dir>")
		}
		runClose(projectDir)
		return
	}

	projectDir := flag.Arg(0)
	if projectDir == "" {
		fmt.Fprintln(os.Stderr, "error: project directory is required")
		fmt.Fprintln(os.Stderr)
		usage()
		os.Exit(1)
	}

	// Resolve to absolute path.
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		fatalf("resolving path: %v", err)
	}

	// Verify the directory exists.
	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		fatalf("%s is not a valid directory", absDir)
	}

	if *resetColor {
		if err := config.Reset(absDir); err != nil {
			fatalf("resetting color: %v", err)
		}
		fmt.Printf("Reset color assignment for %s\n", absDir)
		return
	}

	// Get or assign a color.
	scheme, err := config.GetOrAssign(absDir, *colorName)
	if err != nil {
		fatalf("%v", err)
	}

	projectName := filepath.Base(absDir)
	fmt.Printf("Workspace: %s\n", projectName)
	fmt.Printf("Color:     %s (%s)\n", scheme.Name, scheme.Base)
	fmt.Println()

	// Track all launched processes for session management.
	session := &config.Session{
		ProjectDir: absDir,
		Scheme:     scheme.Name,
		CreatedAt:  time.Now(),
	}

	// Launch Firefox with themed profile.
	if *browser {
		fmt.Println("Setting up Firefox profile...")
		if err := launcher.EnsureFirefoxProfile(scheme); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Firefox profile setup failed: %v\n", err)
		} else {
			fmt.Println("Launching Firefox...")
			proc, err := launcher.LaunchFirefox(scheme)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: Firefox launch failed: %v\n", err)
			} else {
				session.Processes = append(session.Processes, toTracked(proc))
			}
		}
	}

	// Launch Ghostty terminals.
	if !*noTerminals {
		fmt.Printf("Opening %d Ghostty terminal(s)...\n", *terminals)
		procs, err := launcher.LaunchGhostty(scheme, absDir, *terminals)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: Ghostty launch failed: %v\n", err)
		}
		for i := range procs {
			session.Processes = append(session.Processes, toTracked(&procs[i]))
		}
	}

	// Configure and launch Cursor.
	if !*noCursor {
		fmt.Println("Configuring Cursor colors...")
		if err := launcher.ConfigureCursor(scheme, absDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Cursor config failed: %v\n", err)
		}
		fmt.Println("Opening Cursor...")
		proc, err := launcher.LaunchCursor(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: Cursor launch failed: %v\n", err)
		} else {
			session.Processes = append(session.Processes, toTracked(proc))
		}
	}

	// Update JankyBorders window border colour.
	if *borders {
		fmt.Println("Updating JankyBorders...")
		if err := launcher.UpdateBorders(scheme); err != nil {
			fmt.Fprintf(os.Stderr, "warning: JankyBorders update failed: %v\n", err)
		}
	}

	// Save the session so we can close these windows later.
	if len(session.Processes) > 0 {
		if err := config.SaveSession(session); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save session: %v\n", err)
		}
	}

	fmt.Println("Done!")
}

// toTracked converts a launcher.LaunchedProcess to a config.TrackedProcess.
func toTracked(p *launcher.LaunchedProcess) config.TrackedProcess {
	return config.TrackedProcess{
		PID:         p.PID,
		CommandName: p.CommandName,
		Description: p.Description,
		LaunchedAt:  time.Now(),
	}
}

func runClose(projectDir string) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		fatalf("resolving path: %v", err)
	}

	session, err := config.LoadSession(absDir)
	if err != nil {
		fatalf("loading session: %v", err)
	}
	if session == nil {
		fmt.Printf("No active session for %s\n", absDir)
		return
	}

	fmt.Printf("Closing workspace: %s (%s)\n", filepath.Base(absDir), session.Scheme)
	closeSession(session)

	if err := config.DeleteSession(absDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not clean up session file: %v\n", err)
	}
}

func runCloseAll() {
	sessions, err := config.ListSessions()
	if err != nil {
		fatalf("listing sessions: %v", err)
	}
	if len(sessions) == 0 {
		fmt.Println("No active sessions.")
		return
	}

	for _, s := range sessions {
		fmt.Printf("Closing workspace: %s (%s)\n", filepath.Base(s.ProjectDir), s.Scheme)
		closeSession(s)
		if err := config.DeleteSession(s.ProjectDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not clean up session for %s: %v\n", s.ProjectDir, err)
		}
	}
	fmt.Println("All workspaces closed.")
}

func closeSession(s *config.Session) {
	for _, p := range s.Processes {
		if config.IsProcessAlive(p) {
			fmt.Printf("  Closing %s (PID %d)...\n", p.Description, p.PID)
			if err := config.KillProcess(p); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not close %s: %v\n", p.Description, err)
			}
		} else {
			fmt.Printf("  %s (PID %d) already closed\n", p.Description, p.PID)
		}
	}
}

func runList() {
	assignments, err := config.List()
	if err != nil {
		fatalf("listing assignments: %v", err)
	}

	if len(assignments) == 0 {
		fmt.Println("No color assignments yet.")
		fmt.Println()
		fmt.Printf("Available colors: %v\n", color.Names())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tCOLOR\tASSIGNED")
	for dir, a := range assignments {
		fmt.Fprintf(w, "%s\t%s\t%s\n", dir, a.Scheme, a.AssignedAt.Format("2006-01-02 15:04"))
	}
	w.Flush()
}

func usage() {
	fmt.Fprintf(os.Stderr, `workspace — color-coded workspace launcher

Usage:
  workspace <project-dir> [flags]    launch a workspace
  workspace close <project-dir>      close all tracked windows for a project
  workspace --close-all              close all tracked workspace windows
  workspace --list                   list all color assignments

Examples:
  workspace ~/projects/zenml                    # 2 terminals + Cursor
  workspace ~/projects/zenml --terminals 4      # 4 terminals + Cursor
  workspace ~/projects/zenml -c red             # force red color
  workspace ~/projects/zenml --browser           # include Firefox
  workspace ~/projects/zenml --borders           # include JankyBorders
  workspace ~/projects/zenml --no-cursor        # terminals only
  workspace close ~/projects/zenml              # close the workspace
  workspace --close-all                         # close everything
  workspace ~/projects/zenml --reset-color      # unassign color

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Available colors: %v
`, color.Names())
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
