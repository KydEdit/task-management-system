package service

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_Success(t *testing.T) {
	svc := NewAuthService("test-secret")
	token, err := svc.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("the token must not be empty")
	}
}

func TestGenerateToken_UsesHS256(t *testing.T) {
	svc := NewAuthService("test-secret")
	token, err := svc.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, &Claims{})
	if err != nil {
		t.Fatalf("ParseUnverified failed: %v", err)
	}

	hs := parsed.Method.Alg()
	if hs != "HS256" {
		t.Errorf("token alg: expected %q, got %q", "HS256", hs)
	}
}

func TestGenerateToken_ExpirationIs24Hours(t *testing.T) {
	svc := NewAuthService("test-secret")
	token, err := svc.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parser := jwt.NewParser()
	parsed, _, err := parser.ParseUnverified(token, &Claims{})
	if err != nil {
		t.Fatalf("ParseUnverified failed: %v", err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		t.Fatal("failed to cast claims to *Claims")
	}
	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt is nil")
	}
	exp := claims.ExpiresAt.Time
	duration := time.Until(exp)
	// 24h ± 1
	if duration < 24*time.Hour-time.Second || duration > 24*time.Hour+time.Second {
		t.Errorf("exp duration: expected ~24h, got %v", duration)
	}
}

func TestGenerateToken_DifferentSecretsDifferentTokens(t *testing.T) {
	svcA := NewAuthService("secret-a")
	svcB := NewAuthService("secret-b")
	tokenA, err := svcA.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	tokenB, err := svcB.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	partsA := strings.Split(tokenA, ".")
	if len(partsA) != 3 {
		t.Fatalf("tokenA: expected 3 parts, got %d (%q)", len(partsA), tokenA)
	}
	partsB := strings.Split(tokenB, ".")
	if len(partsB) != 3 {
		t.Fatalf("tokenB: expected 3 parts, got %d (%q)", len(partsB), tokenB)
	}

	if partsA[2] == partsB[2] {
		t.Error("signatures must differ for different secrets")
	}
}

func TestParseToken_Success(t *testing.T) {
	svc := NewAuthService("test-secret")
	token, err := svc.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	email, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if email != "test@gmail.com" {
		t.Errorf("ParseToken got email %q, want %q", email, "test@gmail.com")
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	svcA := NewAuthService("secret-a")
	svcB := NewAuthService("secret-b")
	token, err := svcA.GenerateToken("test@gmail.com")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	_, err = svcB.ParseToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	svc := NewAuthService("test-secret")

	claims := &Claims{
		Email: "test@gmail.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = svc.ParseToken(tokenString)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseToken_Malformed(t *testing.T) {
	svc := NewAuthService("test-secret")

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"not a jwt", "not-a-jwt"},
		{"two parts", "aaa.bbb"},
		{"three parts not json", "aaa.bbb.ccc"},
		{"trash", "!@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := svc.ParseToken(tt.token)
			if err == nil {
				t.Errorf("expected error for token %q", tt.token)
			}
			if email != "" {
				t.Errorf("email: expected empty on error, got %q", email)
			}
		})
	}
}

func TestParseToken_AlgNoneRejected(t *testing.T) {
	svc := NewAuthService("test-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{
		Email: "attacker@gmail.com",
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to create alg=none token: %v", err)
	}

	_, err = svc.ParseToken(tokenString)
	if err == nil {
		t.Fatal("expected error for alg=none token")
	}
}

func TestParseToken_WrongAlg(t *testing.T) {
	svc := NewAuthService("test-secret")

	claims := &Claims{
		Email: "test@gmail.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = svc.ParseToken(tokenString)
	if err == nil {
		t.Fatal("expected error for HS512 token (only HS256 allowed)")
	}
}
