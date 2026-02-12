package phone

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

const defaultCountryCode = "ID"

// Normalize converts a phone number to E.164 format.
// It handles Indonesian numbers (081xxx -> +6281xxx) and international format.
func Normalize(phoneNumber string, countryCode string) (string, error) {
	if countryCode == "" {
		countryCode = defaultCountryCode
	}

	// Strip common formatting characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phoneNumber)

	num, err := phonenumbers.Parse(cleaned, countryCode)
	if err != nil {
		return "", err
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", ErrInvalidPhone
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// Validate checks if a phone number is in valid E.164 format.
func Validate(phoneNumber string) bool {
	if !strings.HasPrefix(phoneNumber, "+") {
		return false
	}
	num, err := phonenumbers.Parse(phoneNumber, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(num)
}

// Hash returns a SHA-256 hex hash of a phone number for contact matching.
func Hash(phoneNumber string) string {
	h := sha256.Sum256([]byte(phoneNumber))
	return hex.EncodeToString(h[:])
}
