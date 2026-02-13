package color

// Scheme represents a named workspace color scheme with all the derived
// colors needed by different tools (Ghostty, Cursor, Firefox, JankyBorders).
type Scheme struct {
	Name string `json:"name"`

	// GhosttyBG is the subtle background tint for terminal windows (no # prefix).
	GhosttyBG string `json:"ghostty_bg"`
	// GhosttyFG is the foreground text color for terminals (no # prefix).
	GhosttyFG string `json:"ghostty_fg"`
	// CursorColor is the terminal cursor color (no # prefix).
	CursorColor string `json:"cursor_color"`
	// SelectionBG is the selection background in terminals (no # prefix).
	SelectionBG string `json:"selection_bg"`

	// Accent is the strong color for UI chrome — titlebars, toolbars (with # prefix).
	Accent string `json:"accent"`
	// AccentDim is a dimmed variant for inactive states (with # prefix).
	AccentDim string `json:"accent_dim"`
	// Base is the pure hue at full saturation, used for JankyBorders (with # prefix).
	Base string `json:"base"`
}

// Palettes defines the built-in set of workspace color schemes.
// Ghostty backgrounds are intentionally very subtle — they tint the terminal
// without hurting readability. Accent colors are stronger for UI chrome where
// code readability isn't a concern.
var Palettes = []Scheme{
	{
		Name:        "red",
		GhosttyBG:   "2d1111",
		GhosttyFG:   "f0d0d0",
		CursorColor: "ff6666",
		SelectionBG: "5a2222",
		Accent:      "#6b1a1a",
		AccentDim:   "#4a1111",
		Base:        "#cc3333",
	},
	{
		Name:        "blue",
		GhosttyBG:   "111133",
		GhosttyFG:   "d0d0f0",
		CursorColor: "6666ff",
		SelectionBG: "22225a",
		Accent:      "#1a1a6b",
		AccentDim:   "#11114a",
		Base:        "#3333cc",
	},
	{
		Name:        "green",
		GhosttyBG:   "112d11",
		GhosttyFG:   "d0f0d0",
		CursorColor: "66ff66",
		SelectionBG: "225a22",
		Accent:      "#1a6b1a",
		AccentDim:   "#114a11",
		Base:        "#33cc33",
	},
	{
		Name:        "purple",
		GhosttyBG:   "221133",
		GhosttyFG:   "d0c0f0",
		CursorColor: "9966ff",
		SelectionBG: "33225a",
		Accent:      "#4a1a6b",
		AccentDim:   "#33114a",
		Base:        "#9933cc",
	},
	{
		Name:        "orange",
		GhosttyBG:   "2d2211",
		GhosttyFG:   "f0e0d0",
		CursorColor: "ffaa66",
		SelectionBG: "5a4422",
		Accent:      "#6b4a1a",
		AccentDim:   "#4a3311",
		Base:        "#cc8833",
	},
	{
		Name:        "teal",
		GhosttyBG:   "112d2d",
		GhosttyFG:   "d0f0f0",
		CursorColor: "66ffff",
		SelectionBG: "225a5a",
		Accent:      "#1a6b6b",
		AccentDim:   "#114a4a",
		Base:        "#33cccc",
	},
	{
		Name:        "pink",
		GhosttyBG:   "2d1122",
		GhosttyFG:   "f0d0e0",
		CursorColor: "ff66aa",
		SelectionBG: "5a2244",
		Accent:      "#6b1a4a",
		AccentDim:   "#4a1133",
		Base:        "#cc3399",
	},
	{
		Name:        "gold",
		GhosttyBG:   "2d2d11",
		GhosttyFG:   "f0f0d0",
		CursorColor: "ffff66",
		SelectionBG: "5a5a22",
		Accent:      "#6b6b1a",
		AccentDim:   "#4a4a11",
		Base:        "#cccc33",
	},
}

// ByName returns the scheme with the given name, or nil if not found.
func ByName(name string) *Scheme {
	for i := range Palettes {
		if Palettes[i].Name == name {
			return &Palettes[i]
		}
	}
	return nil
}

// Names returns all available scheme names.
func Names() []string {
	names := make([]string, len(Palettes))
	for i, s := range Palettes {
		names[i] = s.Name
	}
	return names
}
