// Package auth provides password hashing and JWT token utilities.
package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenTTL      = 72 * time.Hour
	bcryptCost    = 12
	jwtSecretEnv  = "JWT_SECRET"
	defaultSecret = "change-me-in-production"
)

// jwtSecret returns the signing secret from the environment, falling back to
// a hardcoded default that is only safe for local development.
func jwtSecret() []byte {
	if s := os.Getenv(jwtSecretEnv); s != "" {
		return []byte(s)
	}
	return []byte(defaultSecret)
}

// HashPassword returns the bcrypt hash of the plain-text password.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword returns nil when plain matches the stored bcrypt hash.
func CheckPassword(plain, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

type claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given user ID valid for tokenTTL.
func GenerateToken(userID int64) (string, error) {
	now := time.Now()
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(jwtSecret())
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return tok, nil
}

// ValidateToken parses and validates a JWT string, returning the user ID
// embedded in the token claims on success.
func ValidateToken(tokenStr string) (int64, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	c, ok := tok.Claims.(*claims)
	if !ok || !tok.Valid {
		return 0, errors.New("invalid token claims")
	}
	return c.UserID, nil
}
