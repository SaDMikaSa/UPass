package color

import (
	"bytes"
	"fmt"
)

const (
	banner = `
    ██╗   ██╗██████╗  █████╗ ███████╗███████╗
    ██║   ██║██╔══██╗██╔══██╗██╔════╝██╔════╝
    ██║   ██║██████╔╝███████║███████╗███████╗
    ██║   ██║██╔═══╝ ██╔══██║╚════██║╚════██║
    ╚██████╔╝██║     ██║  ██║███████║███████║
     ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚══════╝
    `
	subtitle = `
         === Secure Password Manager ===
    `
)

var ColorMap = map[rune]string{
	'█': LightPurple,
	'╗': NavyBlue,
	'╔': NavyBlue,
	'╝': DarkBlue,
	'╚': DarkBlue,
	'║': LightBlue,
	'═': LightBlue,
}

// WelcomeBanner prints the application's welcome banner with or without color formatting.
func WelcomeBanner() {
	fmt.Println(ColorizeBanner(ColorMap) + ApplyColor(subtitle, Aqua+Bold))
}

// SafeColorizeBanner returns a banner with or without colors.
func ColorizeBanner(colorMap map[rune]string) string {
	if !IsColorSupported() {
		return banner
	}

	var result bytes.Buffer
	result.Grow(len(banner) * 4)

	for _, r := range banner {
		if color, ok := colorMap[r]; ok {
			result.WriteString(color)
		}
		result.WriteRune(r)
	}
	result.WriteString(Reset)

	return result.String()
}
