package main

import (
	"math/rand"
)

// Color related constants and functions
const (
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorSilver  = "\033[37m"
	ColorGray    = "\033[90m"
	ColorOrange  = "\033[38;5;214m"
	ColorPurple  = "\033[38;5;93m"
	ColorDate    = "\033[42;30m"
	ColorUser    = "\033[45;97m"
	ColorReset   = "\033[0m"
)

func getRandomColor() string {
	colors := []string{
		ColorRed,
		ColorGreen,
		ColorYellow,
		ColorBlue,
		ColorMagenta,
		ColorCyan,
		ColorSilver,
		ColorGray,
		ColorOrange,
		ColorPurple,
	}

	return colors[rand.Intn(len(colors))]
}
