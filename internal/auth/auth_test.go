package auth_test

import (
	"testing"

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
