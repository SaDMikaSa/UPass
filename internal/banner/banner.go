package banner

import (
	"UPass/internal/checkterm"
	"UPass/internal/crt"
	"fmt"
	"strings"
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
	'█': crt.LightPurple,
	'╗': crt.NavyBlue,
	'╔': crt.NavyBlue,
	'╝': crt.DarkBlue,
	'╚': crt.DarkBlue,
	'║': crt.LightBlue,
	'═': crt.LightBlue,
}

// WelcomeBanner prints the application's welcome banner with or without color formatting.
func WelcomeBanner() {
	fmt.Println(SafeColorizeBanner(ColorMap) + crt.Colorize(subtitle, crt.Aqua+crt.Bold))
}

// SafeColorizeBanner returns a banner with or without colors.
func SafeColorizeBanner(colorMap map[rune]string) string {
	if !checkterm.ColorSupported {
		return banner
	}

	var result strings.Builder
	for _, char := range banner {
		if color, exists := colorMap[char]; exists {
			result.WriteString(color + string(char))
		} else {
			result.WriteString(string(char))
		}
	}
	result.WriteString(crt.Reset)
	return result.String()
}
