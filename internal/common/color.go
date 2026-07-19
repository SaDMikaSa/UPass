package common

import (
	"github.com/fatih/color"
)

var (
	Green  = color.New(color.FgGreen).SprintfFunc()
	Yellow = color.New(color.FgYellow).SprintfFunc()
	Red    = color.New(color.FgRed).SprintfFunc()
	Cyan   = color.New(color.FgCyan).SprintfFunc()
)
