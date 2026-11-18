package security

import (
	"testing"
	"time"
)

func TestHashAndVerifyPassword(t *testing.T) {
	// Arrange
	password := "super-secret-123"

	// Act
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	ok, err := VerifyPassword(password, hash)

	// Assert
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct-horse")
	ok, err := VerifyPassword("battery-staple", hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	if _, err := VerifyPassword("x", "not-a-valid-hash"); err == nil {
		t.Fatal("expected error for malformed hash")
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := "test-secret"
	tok, err := GenerateAccessToken(secret, "user-1", "alice", time.Minute)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	claims, err := ParseAccessToken(secret, tok)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.Subject != "user-1" || claims.Username != "alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestAccessTokenRejectsWrongSecret(t *testing.T) {
	tok, _ := GenerateAccessToken("secret-a", "u", "n", time.Minute)
	if _, err := ParseAccessToken("secret-b", tok); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestAccessTokenExpired(t *testing.T) {
	tok, _ := GenerateAccessToken("s", "u", "n", -time.Minute)
	if _, err := ParseAccessToken("s", tok); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestRefreshTokenHashStable(t *testing.T) {
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if HashRefreshToken(raw) != hash {
		t.Fatal("hash of raw token should match returned hash")
	}
	if len(hash) != 64 {
		t.Fatalf("expected 64-char sha256 hex, got %d", len(hash))
	}
}
