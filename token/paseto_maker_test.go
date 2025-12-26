package token

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestNewPasetoMaker(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "ValidSecretKey",
			secretKey: gofakeit.LetterN(chacha20poly1305.KeySize),
			wantErr:   false,
		},
		{
			name:      "SecretKeyTooShort",
			secretKey: gofakeit.LetterN(chacha20poly1305.KeySize - 1),
			wantErr:   true,
		},
		{
			name:      "SecretKeyTooLong",
			secretKey: gofakeit.LetterN(chacha20poly1305.KeySize + 1),
			wantErr:   true,
		},
		{
			name:      "EmptySecretKey",
			secretKey: "",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maker, err := NewPasetoMaker(tc.secretKey)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != ErrInvalidKeySize {
					t.Errorf("expected ErrInvalidKeySize, got %v", err)
				}
				if maker != nil {
					t.Errorf("expected nil maker, got %v", maker)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if maker == nil {
					t.Errorf("expected maker, got nil")
				}
			}
		})
	}
}

func TestPasetoMaker_CreateToken(t *testing.T) {
	secretKey := gofakeit.LetterN(chacha20poly1305.KeySize)
	maker, err := NewPasetoMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create PASETO maker: %v", err)
	}

	testCases := []struct {
		name     string
		username string
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "ValidToken",
			username: gofakeit.LetterN(10),
			duration: time.Minute,
			wantErr:  false,
		},
		{
			name:     "EmptyUsername",
			username: "",
			duration: time.Minute,
			wantErr:  false,
		},
		{
			name:     "LongUsername",
			username: gofakeit.LetterN(30),
			duration: time.Minute,
			wantErr:  false,
		},
		{
			name:     "ZeroDuration",
			username: gofakeit.LetterN(10),
			duration: 0,
			wantErr:  false,
		},
		{
			name:     "NegativeDuration",
			username: gofakeit.LetterN(10),
			duration: -time.Minute,
			wantErr:  false,
		},
		{
			name:     "LongDuration",
			username: gofakeit.LetterN(10),
			duration: 24 * time.Hour,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := maker.CreateToken(tc.username, tc.duration)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if token == "" {
					t.Errorf("expected non-empty token")
				}
			}
		})
	}
}

func TestPasetoMaker_VerifyToken(t *testing.T) {
	secretKey := gofakeit.LetterN(chacha20poly1305.KeySize)
	maker, err := NewPasetoMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create PASETO maker: %v", err)
	}

	t.Run("ValidToken", func(t *testing.T) {
		username := gofakeit.LetterN(10)
		duration := time.Minute

		token, err := maker.CreateToken(username, duration)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		payload, err := maker.VerifyToken(token)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if payload == nil {
			t.Fatal("expected payload, got nil")
		}
		if payload.Username != username {
			t.Errorf("expected username %s, got %s", username, payload.Username)
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		username := gofakeit.LetterN(10)
		duration := -time.Minute // Already expired

		token, err := maker.CreateToken(username, duration)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		payload, err := maker.VerifyToken(token)
		if err != ErrExpiredToken {
			t.Errorf("expected ErrExpiredToken, got %v", err)
		}
		if payload != nil {
			t.Errorf("expected nil payload for expired token")
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		payload, err := maker.VerifyToken(gofakeit.LetterN(50)) // Random string
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
		if payload != nil {
			t.Errorf("expected nil payload for invalid token")
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		payload, err := maker.VerifyToken("")
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
		if payload != nil {
			t.Errorf("expected nil payload for empty token")
		}
	})

	t.Run("MalformedToken", func(t *testing.T) {
		payload, err := maker.VerifyToken(gofakeit.LetterN(50)) // Random string
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
		if payload != nil {
			t.Errorf("expected nil payload for malformed token")
		}
	})

	t.Run("TokenWithDifferentSecret", func(t *testing.T) {
		differentSecretKey := gofakeit.LetterN(chacha20poly1305.KeySize)
		differentMaker, err := NewPasetoMaker(differentSecretKey)
		if err != nil {
			t.Fatalf("failed to create different PASETO maker: %v", err)
		}

		token, err := differentMaker.CreateToken("testuser", time.Minute)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		// Verify with original maker (different secret)
		payload, err := maker.VerifyToken(token)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
		if payload != nil {
			t.Errorf("expected nil payload for token with different secret")
		}
	})
}

func TestPasetoMaker_TokenPayloadContent(t *testing.T) {
	secretKey := gofakeit.LetterN(chacha20poly1305.KeySize)
	maker, err := NewPasetoMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create PASETO maker: %v", err)
	}

	username := gofakeit.LetterN(10)
	duration := time.Hour

	token, err := maker.CreateToken(username, duration)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	payload, err := maker.VerifyToken(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	// Verify payload contents
	if payload.Username != username {
		t.Errorf("expected username %s, got %s", username, payload.Username)
	}

	if payload.ID.String() == "" {
		t.Error("expected non-empty ID")
	}

	// IssuedAt should be close to now
	if time.Since(payload.IssuedAt) > time.Second {
		t.Errorf("IssuedAt too far from now: %v", payload.IssuedAt)
	}

	// ExpiredAt should be close to IssuedAt + duration
	expectedExpiry := payload.IssuedAt.Add(duration)
	if payload.ExpiredAt.Sub(expectedExpiry) > time.Second {
		t.Errorf("ExpiredAt not matching expected: got %v, expected %v", payload.ExpiredAt, expectedExpiry)
	}
}
