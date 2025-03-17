package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicholasss/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"weakPassword"},
		{"Strong32Passw264!-=d"},
		{"!-v4i!>AFefa86=!31*}"},
	}

	for _, test := range tests {
		actual, err := auth.HashPassword(test.input)
		if err != nil {
			t.Fatalf("Error in HashPassword function: %s", err)
		}

		if auth.CheckPasswordHash(test.input, actual) != nil {
			t.Error("Hash provided does not match.")
		}
	}
}

func TestJWT(t *testing.T) {
	tests := []struct {
		inputUUID   uuid.UUID
		inputSecret string
	}{
		{
			inputUUID:   uuid.New(),
			inputSecret: "new secret",
		},
		{
			inputUUID:   uuid.New(),
			inputSecret: "20f181b52ec1eeb31b2ad0",
		},
	}

	for _, test := range tests {
		duration := 1 * time.Minute

		// create token
		actualToken, err := auth.MakeJWT(test.inputUUID, test.inputSecret, duration)
		if err != nil {
			t.Error("unable to create JWT")
		}

		// compare token
		actualUUID, err := auth.ValidateJWT(actualToken, test.inputSecret)
		if err != nil {
			t.Errorf("unable to validate jwt: %s", err)
		}

		if actualUUID != test.inputUUID {
			t.Errorf("Expected: '%s', Got: '%s'", actualUUID.String(), test.inputUUID.String())
		}
	}
}
