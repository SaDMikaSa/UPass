package checkterm

import (
	"os"
	"strings"
)

var ColorSupported bool

// Init initializes the color support flag.
func Init() {
	ColorSupported = checkColorSupport()
}

// IsColorSupported returns true if the terminal supports color.
func IsColorSupported() bool {
	return ColorSupported
}

// checkColorSupport checks color support through environment variables.
func checkColorSupport() bool {
	if strings.Contains(os.Getenv("TERM"), "color") {
		return true
	}
	//For Unix systems
	if os.Getenv("COLORTERM") != "" {
		return true
	}
	// For Windows
	if os.Getenv("ANSICON") != "" || os.Getenv("OS") == "Windows_NT" {
		return true
	}
	return false
}
