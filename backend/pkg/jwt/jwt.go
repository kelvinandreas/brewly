// Package jwt provides sign and verify helpers for Brewly staff JWTs.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the payload carried by every Brewly JWT.
// Role is populated for access tokens; empty for refresh tokens.
type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role,omitempty"`
	JTI  string `json:"jti"`
}

type rawClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role,omitempty"`
	JTI  string `json:"jti"`
}

// Sign creates a signed HS256 token with the given claims, secret, and TTL.
func Sign(c Claims, secret string, ttl time.Duration) string {
	now := time.Now()
	raw := rawClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.Sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Role: c.Role,
		JTI:  c.JTI,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, raw)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

// Verify parses and validates a token string. Returns Claims on success.
func Verify(tokenStr, secret string) (*Claims, error) {
	raw := &rawClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	if !tok.Valid {
		return nil, ErrInvalid
	}
	return &Claims{Sub: raw.Subject, Role: raw.Role, JTI: raw.JTI}, nil
}

// Sentinel errors returned by Verify.
var (
	ErrExpired = errors.New("jwt: token expired")
	ErrInvalid = errors.New("jwt: token invalid")
)
