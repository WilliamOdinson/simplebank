package token

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTMaker(t *testing.T) {
	testCases := []struct {
		name      string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "ValidSecretKey",
			secretKey: gofakeit.LetterN(32), // 32 chars
			wantErr:   false,
		},
		{
			name:      "SecretKeyTooShort",
			secretKey: "short",
			wantErr:   true,
		},
		{
			name:      "EmptySecretKey",
			secretKey: "",
			wantErr:   true,
		},
		{
			name:      "ExactMinLength",
			secretKey: gofakeit.LetterN(32), // 32 chars
			wantErr:   false,
		},
		{
			name:      "OneBelowMinLength",
			secretKey: gofakeit.LetterN(31), // 31 chars
			wantErr:   true,
		},
		{
			name:      "LongSecretKey",
			secretKey: gofakeit.LetterN(64), // 64 chars
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maker, err := NewJWTMaker(tc.secretKey)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
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

func TestJWTMaker_CreateToken(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	maker, err := NewJWTMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create JWT maker: %v", err)
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

func TestJWTMaker_VerifyToken(t *testing.T) {
	secretKey := gofakeit.LetterN(32)
	maker, err := NewJWTMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create JWT maker: %v", err)
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
		differentSecretKey := gofakeit.LetterN(32)
		differentMaker, err := NewJWTMaker(differentSecretKey)
		if err != nil {
			t.Fatalf("failed to create different JWT maker: %v", err)
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

	t.Run("TokenWithNoneAlgorithm", func(t *testing.T) {
		// Create a token with "none" algorithm (security attack attempt)
		payload, _ := NewPayload(gofakeit.LetterN(10), time.Minute)
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
		token, _ := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)

		result, err := maker.VerifyToken(token)
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken for none algorithm, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil payload for none algorithm token")
		}
	})
}

func TestJWTMaker_TokenPayloadContent(t *testing.T) {
	secretKey := gofakeit.LetterN(32)
	maker, err := NewJWTMaker(secretKey)
	if err != nil {
		t.Fatalf("failed to create JWT maker: %v", err)
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
