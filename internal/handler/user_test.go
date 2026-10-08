package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"task-manager-api/internal/models"
	"testing"
)

type fakeUserService struct {
	meCalled                  bool
	requestedMeEmail          string
	requestedRegisterEmail    string
	requestedRegisterPassword string
	requestedLoginEmail       string
	registerCalled            bool
	loginCalled               bool
	requestedLogin            struct{ email, password string }
	returnedResponse          models.UserResponse
	returnedError             error
}

type fakeAuthService struct {
	requestedEmail      string
	generateTokenCalled bool

	parseTokenCalled bool
	requestedToken   string
	returnedClaims   string

	returnedToken string
	returnedError error
}

func (f *fakeUserService) Login(email string, password string) (models.UserResponse, error) {
	f.loginCalled = true
	f.requestedLoginEmail = email
	f.requestedLogin.email = email
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
	f.requestedRegisterPassword = user.Password

	return f.returnedResponse, f.returnedError
}

func (f *fakeAuthService) GenerateToken(email string) (string, error) {
	f.generateTokenCalled = true
	f.requestedEmail = email

	return f.returnedToken, f.returnedError
}

func (f *fakeAuthService) ParseToken(tokenString string) (string, error) {
	f.parseTokenCalled = true
	f.requestedToken = tokenString

	return f.returnedClaims, f.returnedError
}

func newRequestWithUser(email string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
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
	}
	fakeAuth := &fakeAuthService{}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	req := newRequestWithUser("test@gmail.com")

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
		t.Errorf("requestedMeEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedMeEmail)
	}
	if !fakeService.meCalled {
		t.Error("Me should be called")
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
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
		t.Error("service should not be called on unauthorized request")
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
			req := newRequestWithUser("test@gmail.com")

			h.GetMe(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
			}

			if fakeService.requestedMeEmail != "test@gmail.com" {
				t.Errorf("requestedMeEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedMeEmail)
			}
			if !fakeService.meCalled {
				t.Error("Me should be called")
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
		returnedResponse: models.UserResponse{
			ID:        42,
			Email:     "test@gmail.com",
			CompanyID: 7,
		},
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
	req := httptest.NewRequest(http.MethodPost, "/register", body)

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
		t.Errorf("requestedRegisterEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedRegisterEmail)
	}
	if fakeService.requestedRegisterPassword != "password123" {
		t.Errorf("requestedRegisterPassword: expected %q, got %q", "password123", fakeService.requestedRegisterPassword)
	}
	if !fakeService.registerCalled {
		t.Error("Register should be called")
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
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))

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
			req := httptest.NewRequest(http.MethodPost, "/register", body)

			h.RegisterUser(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("expected %v, got %v", tt.expectedStatus, rec.Code)
			}

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("Content-Type: expected application/json, got %q", ct)
			}

			if fakeService.requestedRegisterEmail != "test@gmail.com" {
				t.Errorf("requestedRegisterEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedRegisterEmail)
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

func TestLoginUser_Success(t *testing.T) {
	fakeService := &fakeUserService{
		returnedResponse: models.UserResponse{
			ID:    42,
			Email: "test@gmail.com",
		},
	}
	fakeAuth := &fakeAuthService{
		returnedToken: "test-token",
	}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	reqBody := models.LoginRequest{
		Email:    "test@gmail.com",
		Password: "password123",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/login", body)

	h.LoginUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["token"] != "test-token" {
		t.Errorf("token: expected %q, got %q", "test-token", resp["token"])
	}

	if fakeService.requestedLoginEmail != "test@gmail.com" {
		t.Errorf("Login requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedLoginEmail)
	}
	if fakeService.requestedLogin.password != "password123" {
		t.Errorf("Login requestedPassword: expected %q, got %q", "password123", fakeService.requestedLogin.password)
	}

	if fakeAuth.requestedEmail != "test@gmail.com" {
		t.Errorf("GenerateToken requestedEmail: expected %q, got %q", "test@gmail.com", fakeAuth.requestedEmail)
	}
	if !fakeAuth.generateTokenCalled {
		t.Error("GenerateToken should be called on successful login")
	}
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	fakeService := &fakeUserService{
		returnedError: models.ErrInvalidCredentials,
	}
	fakeAuth := &fakeAuthService{}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	reqBody := models.LoginRequest{
		Email:    "test@gmail.com",
		Password: "password123",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/login", body)

	h.LoginUser(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if !fakeService.loginCalled {
		t.Error("Login should be called")
	}

	if fakeAuth.generateTokenCalled {
		t.Error("GenerateToken should not be called when credentials are invalid")
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

func TestLoginUser_InvalidJSON(t *testing.T) {
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
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))

			h.LoginUser(rec, req)

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

			if fakeService.loginCalled {
				t.Error("service should not be called for invalid JSON body")
			}

			if fakeAuth.generateTokenCalled {
				t.Error("GenerateToken should not be called for invalid JSON body")
			}
		})
	}
}

func TestLoginUser_TokenGenerationError(t *testing.T) {
	fakeService := &fakeUserService{
		returnedResponse: models.UserResponse{
			ID:    42,
			Email: "test@gmail.com",
		},
	}
	tokenErr := errors.New("failed to sign token")
	fakeAuth := &fakeAuthService{
		returnedError: tokenErr,
	}
	h := NewUserHandler(fakeService, fakeAuth)
	rec := httptest.NewRecorder()
	reqBody := models.LoginRequest{
		Email:    "test@gmail.com",
		Password: "password123",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	body := bytes.NewReader(bodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/login", body)

	h.LoginUser(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: expected application/json, got %q", ct)
	}

	if fakeService.requestedLoginEmail != "test@gmail.com" {
		t.Errorf("Login requestedEmail: expected %q, got %q", "test@gmail.com", fakeService.requestedLoginEmail)
	}
	if fakeService.requestedLogin.password != "password123" {
		t.Errorf("Login requestedPassword: expected %q, got %q", "password123", fakeService.requestedLogin.password)
	}

	if fakeAuth.requestedEmail != "test@gmail.com" {
		t.Errorf("GenerateToken requestedEmail: expected %q, got %q", "test@gmail.com", fakeAuth.requestedEmail)
	}
	if !fakeAuth.generateTokenCalled {
		t.Error("GenerateToken should be called")
	}

	bodyStr := rec.Body.String()
	var respMap map[string]any
	if err := json.Unmarshal([]byte(bodyStr), &respMap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, hasToken := respMap["token"]; hasToken {
		t.Error("response should not contain token field when generation failed")
	}

	errMsg, ok := respMap["error"].(string)
	if !ok || errMsg == "" {
		t.Error("expected error message in response body")
	}
}
