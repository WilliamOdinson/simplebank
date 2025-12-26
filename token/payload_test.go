package token

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
)

func TestNewPayload(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		duration time.Duration
	}{
		{
			name:     "ValidPayload",
			username: gofakeit.LetterN(10),
			duration: time.Minute,
		},
		{
			name:     "EmptyUsername",
			username: "",
			duration: time.Minute,
		},
		{
			name:     "ZeroDuration",
			username: gofakeit.LetterN(10),
			duration: 0,
		},
		{
			name:     "NegativeDuration",
			username: gofakeit.LetterN(10),
			duration: -time.Minute,
		},
		{
			name:     "LongDuration",
			username: gofakeit.LetterN(10),
			duration: 24 * 365 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beforeCreate := time.Now()
			payload, err := NewPayload(tc.username, tc.duration)
			afterCreate := time.Now()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload == nil {
				t.Fatal("expected payload, got nil")
			}

			// Verify ID is set
			if payload.ID.String() == "" || payload.ID.String() == "00000000-0000-0000-0000-000000000000" {
				t.Error("expected valid UUID for ID")
			}

			// Verify username
			if payload.Username != tc.username {
				t.Errorf("expected username %s, got %s", tc.username, payload.Username)
			}

			// Verify IssuedAt is within the creation window
			if payload.IssuedAt.Before(beforeCreate) || payload.IssuedAt.After(afterCreate) {
				t.Errorf("IssuedAt %v not within creation window [%v, %v]", payload.IssuedAt, beforeCreate, afterCreate)
			}

			// Verify ExpiredAt is IssuedAt + duration (within 1ms tolerance for timing)
			expectedExpiry := payload.IssuedAt.Add(tc.duration)
			diff := payload.ExpiredAt.Sub(expectedExpiry)
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Millisecond {
				t.Errorf("expected ExpiredAt close to %v, got %v (diff: %v)", expectedExpiry, payload.ExpiredAt, diff)
			}
		})
	}
}

func TestPayload_UniqueIDs(t *testing.T) {
	ids := make(map[string]bool)
	numPayloads := 100

	for i := 0; i < numPayloads; i++ {
		payload, err := NewPayload(gofakeit.LetterN(10), time.Minute)
		if err != nil {
			t.Fatalf("unexpected error creating payload %d: %v", i, err)
		}
		idStr := payload.ID.String()
		if ids[idStr] {
			t.Errorf("duplicate ID found: %s", idStr)
		}
		ids[idStr] = true
	}
}

func TestPayload_Valid(t *testing.T) {
	t.Run("NotExpired", func(t *testing.T) {
		payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := payload.Valid(); err != nil {
			t.Errorf("expected no error for non-expired token, got %v", err)
		}
	})

	t.Run("Expired", func(t *testing.T) {
		payload, err := NewPayload(gofakeit.LetterN(10), -time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := payload.Valid(); err != ErrExpiredToken {
			t.Errorf("expected ErrExpiredToken, got %v", err)
		}
	})

	t.Run("JustExpired", func(t *testing.T) {
		payload, err := NewPayload(gofakeit.LetterN(10), -time.Nanosecond)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := payload.Valid(); err != ErrExpiredToken {
			t.Errorf("expected ErrExpiredToken, got %v", err)
		}
	})
}

func TestPayload_GetExpirationTime(t *testing.T) {
	payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expTime, err := payload.GetExpirationTime()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if expTime == nil {
		t.Fatal("expected expiration time, got nil")
	}

	expected := jwt.NewNumericDate(payload.ExpiredAt)
	if !expTime.Time.Equal(expected.Time) {
		t.Errorf("expected expiration time %v, got %v", expected.Time, expTime.Time)
	}
}

func TestPayload_GetIssuedAt(t *testing.T) {
	payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issuedAt, err := payload.GetIssuedAt()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if issuedAt == nil {
		t.Fatal("expected issued at time, got nil")
	}

	expected := jwt.NewNumericDate(payload.IssuedAt)
	if !issuedAt.Time.Equal(expected.Time) {
		t.Errorf("expected issued at %v, got %v", expected.Time, issuedAt.Time)
	}
}

func TestPayload_GetNotBefore(t *testing.T) {
	payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	notBefore, err := payload.GetNotBefore()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if notBefore == nil {
		t.Fatal("expected not before time, got nil")
	}

	// NotBefore should equal IssuedAt
	expected := jwt.NewNumericDate(payload.IssuedAt)
	if !notBefore.Time.Equal(expected.Time) {
		t.Errorf("expected not before %v, got %v", expected.Time, notBefore.Time)
	}
}

func TestPayload_GetIssuer(t *testing.T) {
	payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issuer, err := payload.GetIssuer()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if issuer != "SimpleBank Inc." {
		t.Errorf("expected issuer 'SimpleBank Inc.', got '%s'", issuer)
	}
}

func TestPayload_GetSubject(t *testing.T) {
	testCases := []struct {
		name     string
		username string
	}{
		{
			name:     "RegularUsername",
			username: gofakeit.LetterN(10),
		},
		{
			name:     "EmptyUsername",
			username: "",
		},
		{
			name:     "UsernameWithSpecialChars",
			username: gofakeit.LetterN(10) + "!@#$%^&*()",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := NewPayload(tc.username, time.Hour)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			subject, err := payload.GetSubject()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if subject != tc.username {
				t.Errorf("expected subject '%s', got '%s'", tc.username, subject)
			}
		})
	}
}

func TestPayload_GetAudience(t *testing.T) {
	payload, err := NewPayload(gofakeit.LetterN(10), time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	audience, err := payload.GetAudience()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if audience != nil {
		t.Errorf("expected nil audience, got %v", audience)
	}
}

func TestPayload_ImplementsJWTClaims(t *testing.T) {
	payload, _ := NewPayload(gofakeit.LetterN(10), time.Hour)

	// Verify that Payload implements jwt.Claims interface
	var _ jwt.Claims = payload
}
