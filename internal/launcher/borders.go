package launcher

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/strickvl/workspace-colours/internal/color"
)

// UpdateBorders sets the JankyBorders active window border colour to match the
// workspace scheme. JankyBorders is runtime-updatable — if a borders process is
// already running, invoking borders with new options updates it in place.
//
// JankyBorders uses 0xAARRGGBB format (alpha, red, green, blue).
func UpdateBorders(scheme *color.Scheme) error {
	bordersBin, err := exec.LookPath("borders")
	if err != nil {
		return fmt.Errorf("JankyBorders not found in PATH — install with: brew tap FelixKratz/formulae && brew install borders")
	}

	activeColor := hexToJankyBorders(scheme.Base)
	inactiveColor := "0xff333333"

	cmd := exec.Command(bordersBin,
		fmt.Sprintf("active_color=%s", activeColor),
		fmt.Sprintf("inactive_color=%s", inactiveColor),
		"width=6.0",
		"style=round",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("updating borders: %w\n%s", err, string(out))
	}
	return nil
}

// hexToJankyBorders converts a colour like "#cc3333" to JankyBorders format
// "0xffcc3333" (fully opaque).
func hexToJankyBorders(hex string) string {
	hex = strings.TrimPrefix(hex, "#")
	return "0xff" + hex
}
