package util

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestPassword(t *testing.T) {
	password := gofakeit.Password(true, true, true, true, false, 16)

	// Test hashing and verifying password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatal("Failed to hash password:", err)
	}

	err = CheckPassword(password, hashedPassword)
	if err != nil {
		t.Fatal("Password verification failed:", err)
	}

	// Test hashing the same password produces different hashes
	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Fatal("Failed to hash password:", err)
	}

	if hashedPassword == hashedPassword2 {
		t.Fatal("Hashed passwords should not be the same for the same input password")
	}

	// Test wrong password
	wrongPassword := gofakeit.Password(true, true, true, true, false, 17)

	err = CheckPassword(wrongPassword, hashedPassword)
	if err == nil {
		t.Fatal("Password verification should have failed for wrong password")
	}
}
