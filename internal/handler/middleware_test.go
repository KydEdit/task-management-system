package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware_Success(t *testing.T) {
	fakeAuth := &fakeAuthService{returnedClaims: "test@gmail.com"}

	var nextCalled bool
	var capturedEmail string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		capturedEmail, _ = r.Context().Value(userContextKey).(string)
		w.WriteHeader(http.StatusOK)
	})

	h := AuthMiddleware(fakeAuth, next)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	h(rec, req)

	if !nextCalled {
		t.Fatal("next handler should be called")
	}
	if capturedEmail != "test@gmail.com" {
		t.Errorf("email in context: expected %q, got %q", "test@gmail.com", capturedEmail)
	}

	if !fakeAuth.parseTokenCalled {
		t.Error("ParseToken should be called")
	}
	if fakeAuth.requestedToken != "valid-token" {
		t.Errorf("ParseToken got token %q, want %q", fakeAuth.requestedToken, "valid-token")
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	fakeAuth := &fakeAuthService{}

	var nextCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	h := AuthMiddleware(fakeAuth, next)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if nextCalled {
		t.Fatal("next handler should not be called")
	}

	if fakeAuth.parseTokenCalled {
		t.Error("ParseToken should not be called")
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"no token", "Bearer"},
		{"wrong scheme", "Basic abc"},
		{"extra parts", "Bearer abc def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeAuth := &fakeAuthService{}

			var nextCalled bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			h := AuthMiddleware(fakeAuth, next)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			h(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}

			if nextCalled {
				t.Fatal("next should not be called")
			}
			if fakeAuth.parseTokenCalled {
				t.Error("ParseToken should not be called for invalid header format")
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	fakeAuth := &fakeAuthService{returnedError: errors.New("invalid token")}

	var nextCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	h := AuthMiddleware(fakeAuth, next)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	if nextCalled {
		t.Fatal("next should not be called on invalid token")
	}

	if !fakeAuth.parseTokenCalled {
		t.Error("ParseToken should be called")
	}
	if fakeAuth.requestedToken != "invalid-token" {
		t.Errorf("ParseToken got token %q, want %q", fakeAuth.requestedToken, "invalid-token")
	}
}
