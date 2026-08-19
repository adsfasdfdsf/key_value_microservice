package utils

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Verify Hashed Password with Condidate password from form
func VerifyPassword(hashedPassword string, condidate string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(condidate))
	return err == nil
}


// hash existing password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// Validate username to avoid sql injection
func isValidUsername(username string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
    return re.MatchString(username)
}