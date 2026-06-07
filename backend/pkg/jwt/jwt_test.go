package jwt_test

import (
	"testing"
	"time"

	jwtpkg "github.com/kelvinandreas/brewly/pkg/jwt"
)

const (
	testSecret  = "test-secret-32-bytes-long-enough!!"
	otherSecret = "other-secret-32-bytes-long-enough!"
)

func validClaims() jwtpkg.Claims {
	return jwtpkg.Claims{Sub: "user-id-123", Role: "owner", JTI: "jti-abc"}
}

func TestSign_Verify_roundtrip(t *testing.T) {
	tok := jwtpkg.Sign(validClaims(), testSecret, time.Hour)
	got, err := jwtpkg.Verify(tok, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Sub != "user-id-123" {
		t.Errorf("sub: want user-id-123 got %s", got.Sub)
	}
	if got.Role != "owner" {
		t.Errorf("role: want owner got %s", got.Role)
	}
	if got.JTI != "jti-abc" {
		t.Errorf("jti: want jti-abc got %s", got.JTI)
	}
}

func TestVerify_wrongSecret(t *testing.T) {
	tok := jwtpkg.Sign(validClaims(), testSecret, time.Hour)
	_, err := jwtpkg.Verify(tok, otherSecret)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerify_expired(t *testing.T) {
	tok := jwtpkg.Sign(validClaims(), testSecret, -time.Second)
	_, err := jwtpkg.Verify(tok, testSecret)
	if err != jwtpkg.ErrExpired {
		t.Fatalf("want ErrExpired, got %v", err)
	}
}

func TestVerify_tampered(t *testing.T) {
	tok := jwtpkg.Sign(validClaims(), testSecret, time.Hour)
	tampered := tok[:len(tok)-4] + "xxxx"
	_, err := jwtpkg.Verify(tampered, testSecret)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestVerify_refreshToken_noRole(t *testing.T) {
	refresh := jwtpkg.Claims{Sub: "user-id-123", JTI: "refresh-jti"}
	tok := jwtpkg.Sign(refresh, testSecret, 7*24*time.Hour)
	got, err := jwtpkg.Verify(tok, testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Role != "" {
		t.Errorf("role should be empty for refresh token, got %q", got.Role)
	}
}
