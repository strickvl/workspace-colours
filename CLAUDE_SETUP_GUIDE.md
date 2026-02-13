# workspace-colours — Claude Code Setup Guide

This guide is designed to be read by Claude Code (or similar AI assistants) when helping a user set up `workspace-colours` on their machine. It provides structured, step-by-step instructions with verification commands and troubleshooting.

## Overview

`workspace-colours` is a Go CLI tool that assigns persistent colour schemes to project directories and launches colour-coordinated Ghostty terminals, Cursor IDE windows, Firefox browser profiles, and JankyBorders window borders.

**Repository:** This repo contains the Go source code. The binary is called `workspace`.

## Prerequisites Check

Run these checks to determine what's already installed. Each check includes expected output and what to do if the tool is missing.

### 1. Go (required for building)

```bash
go version
```

**Expected:** `go version go1.22+` or higher.
**If missing:** Install via `brew install go` or from https://go.dev/dl/

### 2. Ghostty (required)

```bash
which ghostty || ls /Applications/Ghostty.app/Contents/MacOS/ghostty
```

**Expected:** A path to the ghostty binary.
**If missing:** Install from https://ghostty.org. On macOS, download the .app. The CLI binary is bundled inside the app.

### 3. Cursor (required)

```bash
which cursor
```

**Expected:** A path like `/usr/local/bin/cursor`.
**If missing:** Install Cursor from https://cursor.sh, then open Cursor → Command Palette (Cmd+Shift+P) → type "Install 'cursor' command in PATH" and run it.

### 4. Firefox (optional — for `--browser` flag)

```bash
# Check for Developer Edition (preferred)
ls "/Applications/Firefox Developer Edition.app/Contents/MacOS/firefox" 2>/dev/null && echo "Developer Edition found"

# Check for standard Firefox
ls "/Applications/Firefox.app/Contents/MacOS/firefox-bin" 2>/dev/null && echo "Standard Firefox found"
```

**Expected:** At least one of the above should print a "found" message.
**If missing:** Install from https://www.mozilla.org/firefox/ or https://www.mozilla.org/firefox/developer/
**Note:** Firefox Developer Edition is preferred because it can coexist with standard Firefox and has its own profile system.

### 5. JankyBorders (optional — for `--borders` flag)

```bash
which borders
```

**Expected:** A path like `/usr/local/bin/borders`.
**If missing:** Install with:
```bash
brew tap FelixKratz/formulae && brew install borders
```
**Note:** Requires macOS 14.0+. To auto-start on login: `brew services start borders`

## Building and Installing

### Option A: Install globally via `go install`

```bash
go install github.com/strickvl/workspace-colours/cmd/workspace@latest
```

Verify: `which workspace` should show a path (typically `~/go/bin/workspace`). Make sure `~/go/bin` is in the user's `$PATH`.

### Option B: Build from source

```bash
cd <repo-directory>
go build -o workspace ./cmd/workspace/
```

Then move the binary to somewhere in PATH:
```bash
sudo mv workspace /usr/local/bin/
```

Or add a shell alias in `~/.zshrc`:
```bash
alias workspace='/path/to/repo/workspace'
```

### Verification

```bash
workspace --help
```

**Expected:** Should print usage information with flags like `--terminals`, `--color`, `--browser`, `--borders`, etc.

## First Use Walkthrough

### Step 1: Launch a workspace

Pick any project directory the user works on:

```bash
workspace ~/projects/some-project
```

**What happens:**
1. The tool assigns the first available colour (likely "red") and persists it to `~/.config/workspace-colours/assignments.json`
2. Two Ghostty terminal windows open with a subtle red background tint
3. A `.vscode/settings.json` is created/updated in the project with Cursor colour customizations
4. Cursor opens the project directory

**Verify the colour was saved:**
```bash
workspace --list
```

Should show a table with the project path, colour name, and timestamp.

### Step 2: Try a second project

```bash
workspace ~/projects/another-project
```

This should get a *different* colour (e.g., "blue"), demonstrating the automatic assignment.

### Step 3: Try optional integrations

**Firefox (if installed):**
```bash
workspace ~/projects/some-project --browser --no-cursor --no-terminals
```

This creates a Firefox profile named `workspace-red` with themed toolbar/tabs and launches it.

**IMPORTANT:** The user needs to have `toolkit.legacyUserProfileCustomizations.stylesheets` set to `true` in Firefox's `about:config` for the colour theming to take effect. The tool writes a `user.js` to enable this, but it only takes effect on the *next* Firefox launch with that profile.

**JankyBorders (if installed):**
```bash
workspace ~/projects/some-project --borders --no-cursor --no-terminals
```

The active window border should change to the project's base colour.

### Step 4: Force a specific colour

```bash
workspace ~/projects/some-project --color green
```

This overrides the automatic assignment. The project will be green from now on.

### Step 5: Reset and reassign

```bash
workspace ~/projects/some-project --reset-color
workspace ~/projects/some-project
```

The project will get a new colour assigned (the next available one).

## Configuration Files

The tool creates/uses these files:

| File | Purpose |
|------|---------|
| `~/.config/workspace-colours/assignments.json` | Persistent colour assignments (project path → colour name) |
| `~/.config/ghostty/themes/workspace-*` | Ghostty theme files (one per colour scheme) |
| `<project>/.vscode/settings.json` | Cursor colour customizations (merged into existing settings) |
| `~/Library/Application Support/Firefox/Profiles/*.<name>/chrome/userChrome.css` | Firefox UI theming (one per colour profile) |
| `~/Library/Application Support/Firefox/Profiles/*.<name>/user.js` | Firefox pref to enable userChrome.css |

## Troubleshooting

### "ghostty not found"
The Ghostty binary isn't in PATH. On macOS, ensure `/Applications/Ghostty.app/Contents/MacOS/` is in PATH, or symlink: `ln -s /Applications/Ghostty.app/Contents/MacOS/ghostty /usr/local/bin/ghostty`

### "cursor CLI not found"
Open Cursor → Command Palette → "Install 'cursor' command in PATH". This creates a symlink at `/usr/local/bin/cursor`.

### Firefox colours don't appear
1. Check that the profile was created: `ls ~/Library/Application\ Support/Firefox/Profiles/ | grep workspace-`
2. Check that `userChrome.css` exists in the profile's `chrome/` directory
3. In Firefox, go to `about:config` and verify `toolkit.legacyUserProfileCustomizations.stylesheets` is `true`
4. Restart Firefox with the profile: close Firefox, then run `workspace <dir> --browser` again

### Cursor colours don't change
Check `<project>/.vscode/settings.json` — it should contain a `workbench.colorCustomizations` block. If Cursor was already open, it should live-reload, but try closing and reopening the project.

### JankyBorders not updating
1. Ensure borders is running: `pgrep borders`
2. If not, start it: `brew services start borders`
3. Try running borders directly: `borders active_color=0xffcc3333 width=6.0 style=round`

### Colour assignments are wrong
View the assignments file directly:
```bash
cat ~/.config/workspace-colours/assignments.json
```

To reset everything:
```bash
rm ~/.config/workspace-colours/assignments.json
```

## Available Colour Schemes

| Name | Terminal BG | Accent | Base |
|------|-----------|--------|------|
| red | `#2d1111` | `#6b1a1a` | `#cc3333` |
| blue | `#111133` | `#1a1a6b` | `#3333cc` |
| green | `#112d11` | `#1a6b1a` | `#33cc33` |
| purple | `#221133` | `#4a1a6b` | `#9933cc` |
| orange | `#2d2211` | `#6b4a1a` | `#cc8833` |
| teal | `#112d2d` | `#1a6b6b` | `#33cccc` |
| pink | `#2d1122` | `#6b1a4a` | `#cc3399` |
| gold | `#2d2d11` | `#6b6b1a` | `#cccc33` |

## Architecture (for development)

```
cmd/workspace/main.go          — CLI entry point, flag parsing, orchestration
internal/color/scheme.go       — Colour scheme definitions and lookup
internal/config/config.go      — Assignment persistence (JSON read/write)
internal/launcher/ghostty.go   — Ghostty theme files and window launching
internal/launcher/cursor.go    — Cursor .vscode/settings.json and launching
internal/launcher/firefox.go   — Firefox profile creation and theming
internal/launcher/borders.go   — JankyBorders runtime colour updates
```

Single dependency: `github.com/spf13/pflag` (GNU-style flag parsing).
