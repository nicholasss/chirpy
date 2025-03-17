package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicholasss/chirpy/internal/auth"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		inputToken string
	}{
		{"testing123"},
		{"testing678"},
	}

	for _, test := range tests {
		bearer := "Bearer " + test.inputToken
		headers := http.Header{}
		headers.Set("Authorization", bearer)

		tokenString, err := auth.GetBearerToken(headers)
		if err != nil {
			t.Errorf("Recieved error: '%s'", err)
		}
		if tokenString != test.inputToken {
			t.Errorf("Expected: '%s', Got: '%s'", test.inputToken, tokenString)
		}
	}
}

func TestNoHeadersGetBearerToken(t *testing.T) {
	emptyHeaders := http.Header{}

	_, err := auth.GetBearerToken(emptyHeaders)
	if err == nil {
		t.Errorf("Empty headers should return error")
	}
}

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

func TestEmptyHashPassword(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{""},
	}

	for _, test := range tests {
		_, err := auth.HashPassword(test.input)
		if err == nil {
			t.Fatalf("Expected error from empty password.")
		}
	}
}

func TestNormalJWT(t *testing.T) {
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

func TestExpiredJWT(t *testing.T) {
	tests := []struct {
		inputUUID   uuid.UUID
		inputSecret string
	}{
		{
			inputUUID:   uuid.New(),
			inputSecret: "secret",
		},
		{
			inputUUID:   uuid.New(),
			inputSecret: "secret",
		},
	}

	for _, test := range tests {
		duration := time.Millisecond * 10

		actualToken, err := auth.MakeJWT(test.inputUUID, test.inputSecret, duration)
		if err != nil {
			t.Error("unable to create JWT")
		}

		// sleep 5 second for the JWT to expire
		time.Sleep(time.Millisecond * 30)

		_, err = auth.ValidateJWT(actualToken, test.inputSecret)
		if err == nil {
			t.Error("Expected JWT to expire, causing an error")
		}
	}
}

func TestWrongSecretJWT(t *testing.T) {
	tests := []struct {
		inputUUID        uuid.UUID
		inputSecret      string
		validationSecret string
	}{
		{
			inputUUID:        uuid.New(),
			inputSecret:      "secret1",
			validationSecret: "secret2",
		},
		{
			inputUUID:        uuid.New(),
			inputSecret:      "secret3",
			validationSecret: "secret4",
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
		_, err = auth.ValidateJWT(actualToken, test.validationSecret)
		if err == nil {
			t.Error("Recieved no error for invalid secret with JWT")
		}
	}
}
