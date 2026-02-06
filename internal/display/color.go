package display

import (
	"os"
)

// ANSI color codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

var colorEnabled bool

// OutputFormat controls output mode: "" (terminal), "json", or "md"
var OutputFormat string

func IsJSON() bool { return OutputFormat == "json" }
func IsMD() bool   { return OutputFormat == "md" }

func init() {
	colorEnabled = shouldUseColor()
}

func shouldUseColor() bool {
	// NO_COLOR convention: https://no-color.org/
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// Check if stdout is a terminal
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func apply(code, s string) string {
	if !colorEnabled {
		return s
	}
	return code + s + reset
}

func Bold(s string) string    { return apply(bold, s) }
func Dim(s string) string     { return apply(dim, s) }
func Red(s string) string     { return apply(red, s) }
func Green(s string) string   { return apply(green, s) }
func Yellow(s string) string  { return apply(yellow, s) }
func Blue(s string) string    { return apply(blue, s) }
func Magenta(s string) string { return apply(magenta, s) }
func Cyan(s string) string    { return apply(cyan, s) }
func White(s string) string   { return apply(white, s) }

func BoldCyan(s string) string    { return apply(bold+cyan, s) }
func BoldGreen(s string) string   { return apply(bold+green, s) }
func BoldYellow(s string) string  { return apply(bold+yellow, s) }
func BoldBlue(s string) string    { return apply(bold+blue, s) }
func BoldMagenta(s string) string { return apply(bold+magenta, s) }
func BoldWhite(s string) string   { return apply(bold+white, s) }
