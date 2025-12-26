package token

import (
	"errors"
	"time"
)

var (
	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")

	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("token is invalid")

	// ErrInvalidKeySize is returned when the key size is invalid
	ErrInvalidKeySize = errors.New("invalid key size")
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and valid duration
	CreateToken(username string, duration time.Duration) (string, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
