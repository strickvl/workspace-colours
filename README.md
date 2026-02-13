# workspace-colours

<video src="https://github.com/user-attachments/assets/abfe485c-073a-475a-872f-18b44b98a5a3" width="100%" controls></video>

**One command to launch a color-coded workspace.** Every tool — terminal, IDE, browser — gets the same colour, so you always know which project you're looking at.

Context switching between projects is cognitively expensive when every terminal and editor window looks identical. `workspace-colours` fixes this by assigning a persistent colour to each project directory and applying it across your entire toolchain.

```
$ workspace ~/projects/zenml --terminals 3

Workspace: zenml
Color:     red (#cc3333)

Opening 3 Ghostty terminal(s)...
Configuring Cursor colors...
Opening Cursor...
Done!
```

## What it does

When you run `workspace <project-dir>`:

1. **Assigns a colour** — picks the next unused colour from a palette of 8 schemes, or uses the one already assigned to that project
2. **Opens Ghostty terminals** — launches N terminal windows with a subtle background tint matching the colour
3. **Configures Cursor** — writes `.vscode/settings.json` colour customizations (title bar, activity bar, status bar) and opens the project
4. **Launches Firefox** *(optional)* — opens a Firefox profile themed to match
5. **Updates window borders** *(optional)* — sets JankyBorders to the workspace colour

The colour assignment is persistent — the next time you open the same project, it gets the same colour. All launched windows are tracked, so you can close an entire workspace with one command:

```
$ workspace close ~/projects/zenml

Closing workspace: zenml (red)
  Closing Ghostty — Main (PID 12345)...
  Closing Ghostty — Server (PID 12346)...
  Closing Cursor IDE (PID 12347)...
```

## Supported tools

| Tool | Status | How it works |
|------|--------|-------------|
| [Ghostty](https://ghostty.org) (terminal) | Supported | Custom theme files, launched via CLI flags |
| [Cursor](https://cursor.sh) (IDE) | Supported | Per-workspace `.vscode/settings.json` colour overrides |
| [Firefox](https://www.mozilla.org/firefox/) (browser) | Supported | Separate profiles with themed `userChrome.css` |
| [JankyBorders](https://github.com/FelixKratz/JankyBorders) (window borders) | Supported | Runtime-updatable border colour |

## Prerequisites

- **macOS** (Linux support is possible but untested)
- **[Ghostty](https://ghostty.org)** terminal emulator
- **[Cursor](https://cursor.sh)** IDE with the CLI installed (`Cursor → Command Palette → "Install 'cursor' command in PATH"`)
- **Go 1.22+** (for installation)

**Optional:**
- **[Firefox](https://www.mozilla.org/firefox/)** or **Firefox Developer Edition** — for browser colour theming (activated with `--browser`)
- **[JankyBorders](https://github.com/FelixKratz/JankyBorders)** — for OS-level window border colours (activated with `--borders`). Install: `brew tap FelixKratz/formulae && brew install borders`

## Installation

**Using Claude Code?** Just point it at [`CLAUDE_SETUP_GUIDE.md`](CLAUDE_SETUP_GUIDE.md) — it has step-by-step instructions for checking prerequisites, building, and verifying the install.

```bash
go install github.com/strickvl/workspace-colours/cmd/workspace@latest
```

Or build from source:

```bash
git clone https://github.com/strickvl/workspace-colours.git
cd workspace-colours
go build -o workspace ./cmd/workspace/
# Move the binary somewhere in your PATH
mv workspace /usr/local/bin/
```

## Usage

```bash
# Open 2 terminals + Cursor (default)
workspace ~/projects/zenml

# Open more terminals
workspace ~/projects/zenml --terminals 4
workspace ~/projects/zenml -t 4

# Force a specific colour
workspace ~/projects/zenml --color red
workspace ~/projects/zenml -c blue

# Terminals only (no Cursor)
workspace ~/projects/zenml --no-cursor

# Cursor only (no terminals)
workspace ~/projects/zenml --no-terminals

# List all colour assignments
workspace --list

# Include a themed Firefox window
workspace ~/projects/zenml --browser

# Update JankyBorders window border colour
workspace ~/projects/zenml --borders

# The full experience
workspace ~/projects/zenml -t 3 --browser --borders

# Close all windows for a workspace
workspace close ~/projects/zenml

# Close ALL tracked workspaces
workspace --close-all

# Reset a project's colour
workspace ~/projects/zenml --reset-color
```

Flags can go before or after the project directory — both work:

```bash
workspace ~/projects/zenml --terminals 4
workspace --terminals 4 ~/projects/zenml
```

## Colour palette

Eight named colour schemes are available. Each scheme defines subtle terminal backgrounds (for readability) and stronger accent colours (for UI chrome like title bars).

| Name | Terminal tint | Accent | Base colour |
|------|--------------|--------|-------------|
| red | `#2d1111` | `#6b1a1a` | `#cc3333` |
| blue | `#111133` | `#1a1a6b` | `#3333cc` |
| green | `#112d11` | `#1a6b1a` | `#33cc33` |
| purple | `#221133` | `#4a1a6b` | `#9933cc` |
| orange | `#2d2211` | `#6b4a1a` | `#cc8833` |
| teal | `#112d2d` | `#1a6b6b` | `#33cccc` |
| pink | `#2d1122` | `#6b1a4a` | `#cc3399` |
| gold | `#2d2d11` | `#6b6b1a` | `#cccc33` |

Colours are assigned automatically (first-come, first-served) and persist in `~/.config/workspace-colours/assignments.json`.

## How it works

### Ghostty

For each colour scheme, a Ghostty theme file is written to `~/.config/ghostty/themes/workspace-<name>`. Each terminal window is launched with `--theme=workspace-<name>` and `--working-directory=<project>`. The background tint is intentionally subtle — just enough to recognise the project at a glance without affecting code readability.

### Cursor

Colour customizations are written to `<project>/.vscode/settings.json` under the `workbench.colorCustomizations` key. This tints the title bar, activity bar, status bar, and borders. Existing settings in the file are preserved — only the colour keys are overwritten.

## Configuration

Colour assignments are stored in:

```
~/.config/workspace-colours/assignments.json
```

Session tracking files (for `workspace close`) are stored in:

```
~/.config/workspace-colours/sessions/
```

Ghostty theme files are stored in:

```
~/.config/ghostty/themes/workspace-*
```

### Firefox

When `--browser` is used, a dedicated Firefox profile named `workspace-<colour>` is created (if it doesn't already exist). The profile gets a `userChrome.css` that themes the toolbar, tab bar, and navigation bar, plus a `user.js` that enables the custom stylesheet. Firefox Developer Edition is preferred when available.

### JankyBorders

When `--borders` is used, the JankyBorders active window border colour is updated at runtime to match the workspace. This is a global setting (it can't colour individual windows differently), but it reinforces "what am I working on right now."

## Contributing

Contributions are welcome! Some areas that could use help:

- **Linux support** — different application paths and config locations
- **Additional colour schemes** — the palette is easy to extend

## License

MIT
