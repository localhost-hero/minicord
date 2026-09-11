// Package rooms validates room and participant identifiers before a LiveKit token is issued.
package rooms

import (
	"fmt"
	"regexp"
)

const (
	minNameLength = 3
	maxNameLength = 64
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateRoomName checks that a room name is safe to forward to LiveKit.
func ValidateRoomName(name string) error {
	return validateName("room", name)
}

// ValidateIdentity checks that a participant identity is safe to forward to LiveKit.
func ValidateIdentity(identity string) error {
	return validateName("identity", identity)
}

func validateName(field, value string) error {
	if len(value) < minNameLength || len(value) > maxNameLength {
		return fmt.Errorf("%s must be between %d and %d characters", field, minNameLength, maxNameLength)
	}
	if !namePattern.MatchString(value) {
		return fmt.Errorf("%s must contain only letters, digits, hyphens, and underscores", field)
	}
	return nil
}
