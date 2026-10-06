package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"task-manager-api/internal/models"
	"testing"
)

type fakeUserService struct {
	requestedEmail string

	returnedResponse models.UserResponse
	returnedError    error
}

type fakeAuthService struct {
	requestedEmail string

	returnedToken string

	returnedClaims string

	returnedError error
}

func (f *fakeUserService) Login(email string, password string) (models.UserResponse, error) {

	return f.returnedResponse, f.returnedError
}

func (f *fakeUserService) Me(email string) (models.UserResponse, error) {
	f.requestedEmail = email

	return f.returnedResponse, f.returnedError
}

func (f *fakeUserService) Register(user models.RegisterRequest) (models.UserResponse, error) {

	return f.returnedResponse, f.returnedError
}

func (f *fakeAuthService) GenerateToken(email string) (string, error) {

	return f.returnedToken, f.returnedError
}

func (f *fakeAuthService) ParseToken(tokenString string) (string, error) {

	return f.returnedClaims, f.returnedError
}

func newRequestWithUser(method, target, email string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	ctx := context.WithValue(req.Context(), userContextKey, email)
	return req.WithContext(ctx)
}

func TestGetMe_Success(t *testing.T) {
	req := newRequestWithUser(http.MethodGet, "/get", "test@gmail.com")

	fakeService := &fakeUserService{
		returnedResponse: models.UserResponse{
			ID:        42,
			Email:     "test@gmail.com",
			CompanyID: 7,
		},
	}
	fakeAuth := &fakeAuthService{}

	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp models.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("resp.ID: expected 42, got %d", resp.ID)
	}
	if resp.Email != "test@gmail.com" {
		t.Errorf("resp.Email: expected %q, got %q", "test@gmail.com", resp.Email)
	}
	if resp.CompanyID != 7 {
		t.Errorf("resp.CompanyID: expected 7, got %d", resp.CompanyID)
	}

	if fakeService.requestedEmail != "test@gmail.com" {
		t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedEmail)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	var errResp struct {
		Error string `json:"error"`
	}

	req := httptest.NewRequest(http.MethodGet, "/get", nil)

	fakeService := &fakeUserService{}
	fakeAuth := &fakeAuthService{}

	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	if fakeService.requestedEmail != "" {
		t.Errorf("service should not be called when email is missing from context, but was called with %q", fakeService.requestedEmail)
	}

	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestGetMe_UserNotFound(t *testing.T) {
	var errResp struct {
		Error string `json:"error"`
	}

	req := newRequestWithUser(http.MethodGet, "/get", "test@gmail.com")

	fakeService := &fakeUserService{
		returnedError: models.ErrUserNotFound,
	}
	fakeAuth := &fakeAuthService{}

	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	if fakeService.requestedEmail != "test@gmail.com" {
		t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedEmail)
	}

	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestGetMe_InternalError(t *testing.T) {
	var errResp struct {
		Error string `json:"error"`
	}

	dbErr := errors.New("database unavailable")

	req := newRequestWithUser(http.MethodGet, "/get", "test@gmail.com")

	fakeService := &fakeUserService{
		returnedError: dbErr,
	}
	fakeAuth := &fakeAuthService{}

	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	if fakeService.requestedEmail != "test@gmail.com" {
		t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedEmail)
	}

	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}
