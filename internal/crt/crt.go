package crt

import (
	"UPass/internal/checkterm"
	"fmt"
)

// ANSI color codes
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[38;5;22m"
	Violet      = "\033[38;5;134m"
	DarkBlue    = "\033[38;5;18m"
	Aqua        = "\033[38;5;51m"
	LightPurple = "\033[38;5;141m"
	DarkPurple  = "\033[38;5;55m"
	LightBlue   = "\033[38;5;33m"
	NavyBlue    = "\033[38;5;17m"
	Bold        = "\033[1m"
)

// Colorize applies color to text if support is available.
func Colorize(text string, color string) string {
	return safeColorize(text, color)
}

func Info(text string) {
	if checkterm.IsColorSupported() {
		fmt.Println(Colorize(text, Aqua+Bold))
	} else {
		fmt.Println(text)
	}
}

func Prompts(text string) {
	if checkterm.IsColorSupported() {
		fmt.Print(Colorize("Enter "+text+" --> ", LightBlue))
	} else {
		fmt.Print("Enter " + text + " --> ")
	}
}

func Rejected(text string) {
	if checkterm.IsColorSupported() {
		fmt.Println(Colorize("REJECTED: "+text, Red+Bold))
	} else {
		fmt.Println("REJECTED: " + text)
	}
}

func Success() {
	if checkterm.IsColorSupported() {
		fmt.Println(Colorize("SUCCESS", Green+Bold))
	} else {
		fmt.Println("SUCCESS")
	}
}

func safeColorize(text, color string) string {
	if !checkterm.ColorSupported {
		return text
	}
	return color + text + Reset
}
