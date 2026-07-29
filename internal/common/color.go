package common

import (
	"github.com/fatih/color"
)

func CyanPrintln(a ...interface{}) {
	color.New(color.FgCyan).Println(a...)
}

func CyanPrintf(format string, a ...interface{}) {
	color.New(color.FgCyan).Printf(format, a...)
}

func GreenPrintln(a ...interface{}) {
	color.New(color.FgGreen).Println(a...)
}

func GreenPrintf(format string, a ...interface{}) {
	color.New(color.FgGreen).Printf(format, a...)
}

func RedPrintln(a ...interface{}) {
	color.New(color.FgRed).Println(a...)
}

func RedPrintf(format string, a ...interface{}) {
	color.New(color.FgRed).Printf(format, a...)
}

func YellowPrintln(a ...interface{}) {
	color.New(color.FgYellow).Println(a...)
}

func YellowPrintf(format string, a ...interface{}) {
	color.New(color.FgYellow).Printf(format, a...)
}
