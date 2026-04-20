package auth

import (
	"testing"
	"time"
)

func TestJWTManagerGenerateAndParse(t *testing.T) {
	mgr := NewJWTManager("test_secret", "test_issuer", time.Minute)
	token, _, err := mgr.Generate(123)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	claims, err := mgr.Parse(token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	if claims.UserID != 123 {
		t.Fatalf("unexpected user id: %d", claims.UserID)
	}
}
