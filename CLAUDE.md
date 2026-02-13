# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

`workspace-colours` is a Go CLI tool (`workspace`) that assigns persistent colour schemes to project directories and launches colour-coordinated developer tools: Ghostty terminals, Cursor IDE, Firefox browser profiles, and JankyBorders window borders. macOS-focused; Linux paths exist but are untested.

## Build and Run

```bash
# Build the binary (outputs ./workspace at repo root)
go build -o workspace ./cmd/workspace/

# Run directly
go run ./cmd/workspace/ ~/projects/some-project

# Install globally
go install github.com/strickvl/workspace-colours/cmd/workspace@latest
```

There are no tests yet. There is no linter configuration.

## Architecture

The codebase follows a standard Go layout with a single binary entry point:

```
cmd/workspace/main.go       — CLI entry point: flag parsing with spf13/pflag, then calls launchers
internal/color/scheme.go    — Scheme struct (the central data type) + 8 hardcoded palettes
internal/config/config.go   — Persistent assignment storage (JSON at ~/.config/workspace-colours/assignments.json)
internal/launcher/           — One file per tool integration:
  ghostty.go                — Writes theme files to ~/.config/ghostty/themes/, launches windows via CLI
  cursor.go                 — Merges colour keys into <project>/.vscode/settings.json, launches via `cursor` CLI
  firefox.go                — Creates Firefox profiles, writes userChrome.css + user.js
  borders.go                — Sends runtime colour update to JankyBorders via `borders` CLI
```

**Key design pattern:** `color.Scheme` is the single source of truth. It carries all colour variants (subtle terminal BG, strong accent, pure base hue) and every launcher reads only the fields it needs. Adding a new tool integration means: (1) add any new colour fields to `Scheme`, (2) create a new file in `internal/launcher/`, (3) wire it into `main.go`.

**Colour format inconsistency to be aware of:** Ghostty fields (`GhosttyBG`, `GhosttyFG`, `CursorColor`, `SelectionBG`) omit the `#` prefix. UI chrome fields (`Accent`, `AccentDim`, `Base`) include it. This is intentional — Ghostty config format doesn't use `#`.

## Dependencies

Single external dependency: `github.com/spf13/pflag` (GNU-style CLI flags). Everything else is stdlib.

## Config Files the Tool Creates

- `~/.config/workspace-colours/assignments.json` — project-path-to-colour mappings
- `~/.config/ghostty/themes/workspace-*` — one theme file per colour scheme
- `<project>/.vscode/settings.json` — Cursor colour customizations (merged, not overwritten)
- Firefox profile dirs under `~/Library/Application Support/Firefox/Profiles/` — `userChrome.css` + `user.js`

## Conventions

- British spelling for user-facing text ("colour"), American spelling in Go identifiers (`color`) to match Go stdlib
- All launcher functions are non-blocking for GUI apps (use `cmd.Start()`, not `cmd.Run()`)
- Errors from optional tools (Firefox, JankyBorders) are printed as warnings, not fatal
- The `design/` folder is in `.gitignore` — never commit its contents
