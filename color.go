package main

import (
	"math/rand"
	"time"
)

const (
	ColorReset           = "\033[0m"
	ColorWhiteText       = "\033[1;97m"
	ColorNotification    = "\033[5;92m"
	ColorWhiteBackground = "\033[47m"
)

var random *rand.Rand

func init() {
	// Seed the random number generator once at startup
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func getRandomColor() string {
	colors := []string{
		"\033[0;104m",
		"\033[0;105m",
		"\033[0;106m",
		"\033[1;100m",
		"\033[1;103m",
		"\033[1;41m",
		"\033[1;42m",
	}

	return colors[random.Intn(len(colors))]
}
