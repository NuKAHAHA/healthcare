package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager handles password hashing and verification
type PasswordManager struct {
	// bcrypt cost factor - higher = slower but more secure
	// 12 is a good balance for modern hardware
	cost int
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		cost: 12,
	}
}

// HashPassword hashes a password using bcrypt
// Returns hash or error if hashing fails
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
// Returns true if password matches, false otherwise
func (pm *PasswordManager) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
