package common

import (
	"github.com/fatih/color"
)

func Green(format string, a ...interface{}) string {
	return color.GreenString(format, a...)
}

func Yellow(format string, a ...interface{}) string {
	return color.YellowString(format, a...)
}

func Red(format string, a ...interface{}) string {
	return color.RedString(format, a...)
}

func Cyan(format string, a ...interface{}) string {
	return color.CyanString(format, a...)
}
