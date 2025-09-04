package color

import (
	"os"
	"runtime"

	"golang.org/x/term"
)

// SupportedTerms is a set of known terminal types that support color output.
// If a terminal type is in this map, color support is assumed.
var (
	SupportedTerms = map[string]struct{}{
		"xterm":           {},
		"screen":          {},
		"tmux":            {},
		"linux":           {},
		"alacritty":       {},
		"kitty":           {},
		"wezterm":         {},
		"xterm-256color":  {},
		"screen-256color": {},
	}
	colorEnabled bool
)

// Init initializes color support detection by checking terminal capabilities.
func Init() {
	colorEnabled = checkColorSupport()
}

// IsColorSupported reports whether color output is enabled based on terminal capabilities.
// Returns true if color is supported, false otherwise.
func IsColorSupported() bool {
	return colorEnabled
}

// checkColorSupport performs the logic of determining color support.
func checkColorSupport() bool {

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if !isTerminal(os.Stdout) {
		return false
	}

	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	if _, ok := SupportedTerms[term]; ok {
		return true
	}

	colorterm := os.Getenv("COLORTERM")
	if colorterm != "" && (colorterm == "truecolor" || colorterm == "24bit") {
		return true
	}

	if runtime.GOOS == "windows" {
		return isWindowsColorSupported()
	}

	return false
}

// isTerminal checks whether the transferred file descriptor is a terminal.
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// isWindowsColorSupported implements Windows-specific logic.
func isWindowsColorSupported() bool {
	// Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		return true
	}
	// ANSICON for older versions
	if os.Getenv("ANSICON") != "" {
		return true
	}
	// ConEmu, Cmder
	if os.Getenv("ConEmuANSI") == "ON" {
		return true
	}
	return false
}
