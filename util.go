package main

import (
	"regexp"
)

const (
	minUsernameLength = 5
	maxUsernameLength = 8

	minRoomNameLength = 5
	maxRoomNameLength = 20
)

// isValidUsername checks if the username is valid (3-15 characters, alphanumeric + _).
func isValidUsername(username string) bool {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return false
	}
	validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validName.MatchString(username)
}

// isValidRoomName checks if the room name is valid (3-20 characters, alphanumeric + _).
func isValidRoomName(roomName string) bool {
	if len(roomName) < minRoomNameLength || len(roomName) > maxRoomNameLength {
		return false
	}
	validRoom := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validRoom.MatchString(roomName)
}
