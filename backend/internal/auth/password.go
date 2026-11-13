// Package auth provides password hashing and token helpers for authentication.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash of plain suitable for storage.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
