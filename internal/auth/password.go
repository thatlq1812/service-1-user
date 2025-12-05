package auth

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// ValidatePassword check if the password meets strength requirements
// Require at least 8 characters, one lowercase, one uppercase, and one digit (alphanumeric only)
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasDigit   = regexp.MustCompile(`\d`).MatchString(password)
		validChars = regexp.MustCompile(`^[a-zA-Z\d]+$`).MatchString(password)
	)

	if !hasLower || !hasUpper || !hasDigit {
		return errors.New("password must contain at least one lowercase letter, one uppercase letter, and one digit")
	}

	if !validChars {
		return errors.New("password must only contain alphanumeric characters")
	}

	return nil
}

// HashPassword generates a bcrypt hash from a plain text password
// Uses bcrypt.DefaultCost (currently 10) for hashing strength
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}

// CheckPassword compares a plain text password with a bcrypt hashed password
// Returns true if the password matches the hash, false otherwise
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
