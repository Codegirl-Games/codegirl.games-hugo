package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Authenticator verifies admin credentials.
type Authenticator struct {
	username     string
	passwordHash string
}

// New creates an Authenticator with the given credentials.
func New(username, passwordHash string) *Authenticator {
	return &Authenticator{
		username:     username,
		passwordHash: passwordHash,
	}
}

// Verify checks username and password against stored credentials.
func (a *Authenticator) Verify(username, password string) error {
	if username != a.username {
		return errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(a.passwordHash), []byte(password)); err != nil {
		return errors.New("invalid credentials")
	}

	return nil
}

// HashPassword generates a bcrypt hash suitable for ADMIN_PASSWORD_HASH.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
