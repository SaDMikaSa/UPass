package color

import (
	"fmt"
)

// ANSI color codes
const (
	// Primary colors
	Reset  = "\033[0m"
	Violet = "\033[38;5;134m"
	Red    = "\033[31m"
	Green  = "\033[38;5;22m"
	Aqua   = "\033[38;5;51m"
	// Text style
	Bold = "\033[1m"
	// Combined colors
	RejectColor  = Red + Bold
	SuccessColor = Green + Bold
	InfoColor    = Aqua + Bold
	// Color codes for banner
	DarkBlue    = "\033[38;5;18m"
	LightPurple = "\033[38;5;141m"
	DarkPurple  = "\033[38;5;55m"
	LightBlue   = "\033[38;5;33m"
	NavyBlue    = "\033[38;5;17m"
	// Messages
	enterPromt   = "Enter "
	arrow        = " --> "
	rejectLabel  = "REJECTED: "
	successLabel = "SUCCESS"
)

// applyColor applies color to text supported by the terminal.
func ApplyColor(text string, color string) string {
	if color == "" || !IsColorSupported() {
		return text
	}
	return color + text + Reset
}

// Info prints informational message.
func PrintInfo(text string) {
	fmt.Println(ApplyColor(text, InfoColor))
}

// Prompts prints prompt text.
func PrintPrompts(text string) {
	fmt.Print(ApplyColor(enterPromt+text+arrow, Violet))
}

// Rejected prints rejection message.
func PrintRejected(text string) {
	fmt.Println(ApplyColor(rejectLabel+text, RejectColor))
}

// Success prints success message.
func PrintSuccess() {
	fmt.Println(ApplyColor(successLabel, SuccessColor))
}
