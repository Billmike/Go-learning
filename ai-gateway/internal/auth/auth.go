package auth

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
)

var (
	// ErrMissingToken is returned when the Authorization header is absent.
	ErrMissingToken = errors.New("missing authorization token")

	// ErrInvalidToken is returned when the bearer token is malformed or wrong.
	ErrInvalidToken = errors.New("invalid authorization token")
)

const bearerPrefix = "Bearer "

// ValidateBearer checks that the request carries a valid Bearer token.
// Token comparison uses constant-time equality to avoid timing leaks.
func ValidateBearer(r *http.Request, expected string) error {
	token, err := extractBearer(r)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
		return ErrInvalidToken
	}

	return nil
}

// ExtractToken returns the bearer token from the Authorization header,
// or an empty string if the header is missing or malformed.
func ExtractToken(r *http.Request) string {
	token, err := extractBearer(r)
	if err != nil {
		return ""
	}
	return token
}

func extractBearer(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", ErrInvalidToken
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}
