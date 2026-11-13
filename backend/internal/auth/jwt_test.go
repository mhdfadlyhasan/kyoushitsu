package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParseToken(t *testing.T) {
	token, err := IssueToken("secret", 42, "teacher", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.Subject != "42" {
		t.Fatalf("subject = %q, want 42", claims.Subject)
	}
	if claims.Role != "teacher" {
		t.Fatalf("role = %q, want teacher", claims.Role)
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	token, err := IssueToken("secret", 1, "student", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if _, err := ParseToken("other-secret", token); err == nil {
		t.Fatal("expected token signed with a different secret to fail")
	}
}

func TestParseTokenExpired(t *testing.T) {
	token, err := IssueToken("secret", 1, "student", -time.Minute)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if _, err := ParseToken("secret", token); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func TestParseTokenRejectsUnexpectedAlgorithm(t *testing.T) {
	claims := Claims{
		Role: "teacher",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := ParseToken("secret", token); err == nil {
		t.Fatal("expected non-HS256 token to be rejected")
	}
}

func TestParseTokenTampered(t *testing.T) {
	token, err := IssueToken("secret", 1, "student", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := ParseToken("secret", tampered); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}
