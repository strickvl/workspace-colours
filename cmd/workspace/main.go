package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

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
	flag.Usage = usage

	flag.Parse()

	if *list {
		runList()
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

	// Launch Firefox with themed profile.
	if *browser {
		fmt.Println("Setting up Firefox profile...")
		if err := launcher.EnsureFirefoxProfile(scheme); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Firefox profile setup failed: %v\n", err)
		} else {
			fmt.Println("Launching Firefox...")
			if err := launcher.LaunchFirefox(scheme); err != nil {
				fmt.Fprintf(os.Stderr, "warning: Firefox launch failed: %v\n", err)
			}
		}
	}

	// Launch Ghostty terminals.
	if !*noTerminals {
		fmt.Printf("Opening %d Ghostty terminal(s)...\n", *terminals)
		if err := launcher.LaunchGhostty(scheme, absDir, *terminals); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Ghostty launch failed: %v\n", err)
		}
	}

	// Configure and launch Cursor.
	if !*noCursor {
		fmt.Println("Configuring Cursor colors...")
		if err := launcher.ConfigureCursor(scheme, absDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Cursor config failed: %v\n", err)
		}
		fmt.Println("Opening Cursor...")
		if err := launcher.LaunchCursor(absDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: Cursor launch failed: %v\n", err)
		}
	}

	fmt.Println("Done!")
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
  workspace <project-dir> [flags]

Examples:
  workspace ~/projects/zenml                    # 2 terminals + Cursor
  workspace ~/projects/zenml --terminals 4      # 4 terminals + Cursor
  workspace ~/projects/zenml -c red             # force red color
  workspace ~/projects/zenml --browser           # include Firefox
  workspace ~/projects/zenml --no-cursor        # terminals only
  workspace --list                              # show all assignments
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
