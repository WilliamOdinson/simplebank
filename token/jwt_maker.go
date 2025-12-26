package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// MIN_SECRET_KEY_SIZE is the minimum size of the secret key
	MIN_SECRET_KEY_SIZE = 32
)

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// CreateToken creates a new token for a specific username and valid duration
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < MIN_SECRET_KEY_SIZE {
		return nil, fmt.Errorf("invalid secret key size: must be at least 32 characters")
	}

	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", fmt.Errorf("failed to create payload: %w", err)
	}

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(maker.secretKey))

	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return jwtToken, nil
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)

	if errors.Is(err, jwt.ErrTokenExpired) { // check for expired token
		return nil, ErrExpiredToken
	} else if err != nil {
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)

	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
