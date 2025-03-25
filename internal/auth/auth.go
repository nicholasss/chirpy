package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// utility functions

func GetAPIKey(headers http.Header) (string, error) {
	// value will look like:
	//   ApiKey <key string>

	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("header field 'authorization' is absent")
	}

	keyString, ok := strings.CutPrefix(authHeader, "ApiKey ")
	if !ok {
		log.Printf("Unable to cut prefix off. Before: '%s' After: '%s'", authHeader, keyString)
		return "", errors.New("unable to find key in headers")
	}

	return keyString, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	// value will look like
	//   Bearer <token_string>

	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("header field 'authorization' is absent")
	}

	tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		log.Printf("Unable to cut prefix off. Before: '%s' After: '%s'", authHeader, tokenString)
		return "", errors.New("unable to find token in headers")
	}

	return tokenString, nil
}

// password functions

func HashPassword(password string) (string, error) {
	if password == "" {
		log.Print("Empty password provided.")
		return "", fmt.Errorf("unable to hash empty password")
	}

	hashedData, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Unable to hash password: %s", err)
		return "", err
	}

	return string(hashedData), nil
}

// password is from a request, hash is from the db
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("Unable to compare hash and password: %s", err)
		return err
	}

	return nil
}

// refresh tokens

func MakeRefreshToken() (string, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}

	secureString := hex.EncodeToString(data)
	return secureString, nil
}

// JWT tokens

// creates and returns a JWT
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	currentTime := time.Now().UTC()
	expirationTime := currentTime.UTC().Add(expiresIn)

	signingMethod := jwt.SigningMethodHS256
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(signingMethod, claims)

	// HMAC signing method requires the type []byte
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("Error signing JWT: %s", err)
		return "", err
	}

	// log.Printf("Generated new token: %s", signedToken)
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return uuid.Nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		})
	if err != nil {
		// log.Printf("Error validating JWT: %v", err)
		return uuid.Nil, err
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		// log.Printf("Error accessing token subject: %v", err)
		return uuid.Nil, err
	}
	// log.Printf("UserID from token: %s", userID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		// log.Printf("Error parsing userID from token claim: %v", err)
		return uuid.Nil, err
	}

	return userUUID, nil
}
