package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-manager-api/internal/models"
	"testing"
)

type fakeUserService struct {
	meCalled               bool
	requestedMeEmail       string
	requestedRegisterEmail string
	requestedLoginEmail    string
	registerCalled         bool
	loginCalled            bool
	requestedLogin         struct{ email, password string }
	returnedID             int
	returnedResponse       models.UserResponse
	returnedError          error
}

type fakeAuthService struct {
	requestedEmail string

	returnedToken string

	returnedClaims string

	returnedError error
}

func (f *fakeUserService) Login(email string, password string) (models.UserResponse, error) {
	f.loginCalled = true
	f.requestedLoginEmail = email
	f.requestedLogin.password = password

	return f.returnedResponse, f.returnedError
}

func (f *fakeUserService) Me(email string) (models.UserResponse, error) {
	f.meCalled = true
	f.requestedMeEmail = email

	return f.returnedResponse, f.returnedError
}

func (f *fakeUserService) Register(user models.RegisterRequest) (models.UserResponse, error) {
	f.registerCalled = true
	f.requestedRegisterEmail = user.Email

	f.returnedResponse.ID = f.returnedID
	f.returnedResponse.Email = user.Email
	f.returnedResponse.CompanyID = user.CompanyID

	return f.returnedResponse, f.returnedError
}

func (f *fakeAuthService) GenerateToken(email string) (string, error) {

	return f.returnedToken, f.returnedError
}

func (f *fakeAuthService) ParseToken(tokenString string) (string, error) {

	return f.returnedClaims, f.returnedError
}

func newRequestWithUser(method, target, email string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	ctx := context.WithValue(req.Context(), userContextKey, email)
	return req.WithContext(ctx)
}

func TestGetMe_Success(t *testing.T) {
	fakeService := &fakeUserService{
		returnedResponse: models.UserResponse{
			ID:        42,
			Email:     "test@gmail.com",
			CompanyID: 7,
		},
		returnedID: 42,
	}
	fakeAuth := &fakeAuthService{}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	req := newRequestWithUser(http.MethodGet, "/get", "test@gmail.com", nil)

	h.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
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

	if fakeService.requestedMeEmail != "test@gmail.com" {
		t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedMeEmail)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	var errResp struct {
		Error string `json:"error"`
	}
	fakeService := &fakeUserService{}
	fakeAuth := &fakeAuthService{}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get", nil)

	h.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if fakeService.meCalled {
		t.Errorf("service should not be called, but was called with %q", fakeService.requestedMeEmail)
	}

	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response body")
	}
}

func TestGetMe_Errors(t *testing.T) {
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "user not found",
			serviceError:   models.ErrUserNotFound,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal error",
			serviceError:   errors.New("database unavailable"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &fakeUserService{returnedError: tt.serviceError}
			fakeAuth := &fakeAuthService{}
			h := NewUserHandler(fakeService, fakeAuth)
			rec := httptest.NewRecorder()
			req := newRequestWithUser(http.MethodGet, "/get", "test@gmail.com", nil)

			h.GetMe(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
			}

			if fakeService.requestedMeEmail != "test@gmail.com" {
				t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedMeEmail)
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
		})
	}
}

func TestRegisterUser_Success(t *testing.T) {
	fakeService := &fakeUserService{
		returnedID: 42,
	}
	fakeAuth := &fakeAuthService{}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	reqBody := models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := newRequestWithUser(http.MethodPost, "/register", "test@gmail.com", body)

	h.RegisterUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
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

	if fakeService.requestedRegisterEmail != "test@gmail.com" {
		t.Errorf("requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedRegisterEmail)
	}
}

func TestRegisterUser_InvalidBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed json", "{invalid}"},
		{"wrong type", `{"email": 123}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &fakeUserService{}
			fakeAuth := &fakeAuthService{}
			h := NewUserHandler(fakeService, fakeAuth)
			rec := httptest.NewRecorder()
			req := newRequestWithUser(http.MethodPost, "/register", "test@gmail.com", strings.NewReader(tt.body))

			h.RegisterUser(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
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

			if fakeService.registerCalled {
				t.Error("service should not be called for invalid JSON body")
			}
		})
	}
}

func TestRegisterUser_Errors(t *testing.T) {
	validBody := models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	}
	bodyBytes, err := json.Marshal(validBody)
	if err != nil {
		t.Fatalf("failed to marshal valid body: %v", err)
	}

	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "invalid email",
			serviceError:   models.ErrInvalidEmail,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid company",
			serviceError:   models.ErrInvalidCompanyID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid password",
			serviceError:   models.ErrInvalidPassword,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "duplicate email",
			serviceError:   models.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "company not found",
			serviceError:   models.ErrCompanyNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "could not create user",
			serviceError:   errors.New("database unavailable"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeService := &fakeUserService{
				returnedError: tt.serviceError,
			}
			fakeAuth := &fakeAuthService{}
			h := NewUserHandler(fakeService, fakeAuth)
			rec := httptest.NewRecorder()
			body := bytes.NewReader(bodyBytes)
			req := newRequestWithUser(http.MethodPost, "/register", "test@gmail.com", body)

			h.RegisterUser(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
			}

			if !fakeService.registerCalled {
				t.Error("Register should be called")
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
		})
	}
}
