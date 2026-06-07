// Package tabletoken provides sign and verify helpers for Brewly customer table tokens.
// These are distinct from staff JWTs: they carry tid/tvr claims and have no role.
package tabletoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ttl = 4 * time.Hour

// TableClaims is the payload of a customer table token.
type TableClaims struct {
	TableID      string `json:"tid"`
	TokenVersion int    `json:"tvr"`
	JTI          string `json:"jti"`
}

type rawTableClaims struct {
	jwt.RegisteredClaims
	TableID      string `json:"tid"`
	TokenVersion int    `json:"tvr"`
	JTI          string `json:"jti"`
}

// Sign creates a signed HS256 table token valid for 4 hours.
func Sign(tableID uuid.UUID, tokenVersion int, secret string) string {
	now := time.Now()
	raw := rawTableClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		TableID:      tableID.String(),
		TokenVersion: tokenVersion,
		JTI:          uuid.New().String(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, raw)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

// Verify parses and validates a table token string. Returns TableClaims on success.
// Note: it does NOT check token_version against the DB — that is the middleware's job.
func Verify(tokenStr, secret string) (*TableClaims, error) {
	raw := &rawTableClaims{}
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
	return &TableClaims{
		TableID:      raw.TableID,
		TokenVersion: raw.TokenVersion,
		JTI:          raw.JTI,
	}, nil
}

// Sentinel errors returned by Verify.
var (
	ErrExpired = errors.New("tabletoken: token expired")
	ErrInvalid = errors.New("tabletoken: token invalid")
)
