package service

import (
	"errors"
	"strings"
	"task-manager-api/internal/models"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	registeredEmail     string
	registeredPassword  string
	registeredCompanyID int

	registerCalled bool

	requestedEmail string

	returnedUser  models.User
	returnedID    int
	returnedError error
}

type fakeCompanyRepository struct {
	checkedCompanyID int

	ensureCalled bool

	returnedError error
}

func (f *fakeUserRepository) RegisterUser(email, password string, companyID int) (int, error) {
	f.registerCalled = true

	f.registeredEmail = email
	f.registeredPassword = password
	f.registeredCompanyID = companyID

	return f.returnedID, f.returnedError
}

func (f *fakeUserRepository) GetByEmail(email string) (models.User, error) {
	f.requestedEmail = email

	return f.returnedUser, f.returnedError
}

func (f *fakeCompanyRepository) EnsureExists(companyID int) error {
	f.ensureCalled = true

	f.checkedCompanyID = companyID

	return f.returnedError
}

func TestRegisterUser_Success(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{returnedID: 42}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	resp, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	if fakeCompanyRepo.checkedCompanyID != 7 {
		t.Errorf("fakeCompanyRepo.checkedCompanyID: expected 7, got %d", fakeCompanyRepo.checkedCompanyID)
	}

	if fakeUserRepo.registeredPassword == "password123" {
		t.Errorf("password should be hashed, but got plain: %q", fakeUserRepo.registeredPassword)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(fakeUserRepo.registeredPassword),
		[]byte("password123"),
	)
	if err != nil {
		t.Fatalf("stored password is not a valid hash of the original password: %v", err)
	}
}

func TestRegisterUser_CompanyNotFound(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{}
	fakeCompanyRepo := &fakeCompanyRepository{returnedError: models.ErrCompanyNotFound}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, models.ErrCompanyNotFound) {
		t.Errorf("expected ErrCompanyNotFound, got %v", err)
	}

	if fakeUserRepo.registerCalled {
		t.Error("RegisterUser should not be called when company does not exist")
	}
}

func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{returnedError: models.ErrUserAlreadyExists}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !fakeUserRepo.registerCalled {
		t.Error("RegisterUser was not called, service failed before reaching repository")
	}

	if !errors.Is(err, models.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestRegisterUser_RepositoryError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	fakeUserRepo := &fakeUserRepository{returnedError: dbErr}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 7,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !fakeUserRepo.registerCalled {
		t.Error("RegisterUser was not called, service failed before reaching repository")
	}

	if !errors.Is(err, dbErr) {
		t.Errorf("expected %v, got %v", dbErr, err)
	}
}

func TestRegisterUser_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     models.RegisterRequest
		expectedErr error
	}{
		{
			name: "invalid email",
			request: models.RegisterRequest{
				Email:     "test@",
				Password:  "password123",
				CompanyID: 7,
			},
			expectedErr: models.ErrInvalidEmail,
		},
		{
			name: "empty email",
			request: models.RegisterRequest{
				Email:     "",
				Password:  "password123",
				CompanyID: 7,
			},
			expectedErr: models.ErrInvalidEmail,
		},
		{
			name: "short password",
			request: models.RegisterRequest{
				Email:     "test@gmail.com",
				Password:  "pass",
				CompanyID: 7,
			},
			expectedErr: models.ErrInvalidPassword,
		},
		{
			name: "password too long",
			request: models.RegisterRequest{
				Email:     "test@gmail.com",
				Password:  strings.Repeat("a", 73),
				CompanyID: 7,
			},
			expectedErr: models.ErrInvalidPassword,
		},
		{
			name: "invalid company",
			request: models.RegisterRequest{
				Email:     "test@gmail.com",
				Password:  "password123",
				CompanyID: 0,
			},
			expectedErr: models.ErrInvalidCompanyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUserRepo := &fakeUserRepository{}
			fakeCompanyRepo := &fakeCompanyRepository{}
			svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

			_, err := svc.Register(tt.request)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected %v, got %v", tt.expectedErr, err)
			}

			if fakeUserRepo.registerCalled {
				t.Errorf("RegisterUser should not be called for case %q", tt.name)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	hashedPassword := string(hashed)

	returnedUser := models.User{
		ID:           42,
		Email:        "test@gmail.com",
		PasswordHash: hashedPassword,
	}

	fakeUserRepo := &fakeUserRepository{
		returnedUser: returnedUser,
	}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	resp, err := svc.Login("test@gmail.com", "password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("resp.ID: expected 42, got %d", resp.ID)
	}

	if resp.Email != "test@gmail.com" {
		t.Errorf("resp.Email: expected %q, got %q", "test@gmail.com", resp.Email)
	}

	if fakeUserRepo.requestedEmail != "test@gmail.com" {
		t.Errorf("repo requestedEmail: expected %q, got %q", "test@gmail.com", fakeUserRepo.requestedEmail)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{
		returnedError: models.ErrUserNotFound,
	}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Login("test@gmail.com", "password123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if fakeUserRepo.requestedEmail != "test@gmail.com" {
		t.Errorf("repo requestedEmail: expected %q, got %q", "test@gmail.com", fakeUserRepo.requestedEmail)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	returnedUser := models.User{
		ID:           42,
		Email:        "test@gmail.com",
		PasswordHash: string(hashed),
	}

	fakeUserRepo := &fakeUserRepository{
		returnedUser: returnedUser,
	}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err = svc.Login("test@gmail.com", "wrong-password")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, models.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if fakeUserRepo.requestedEmail != "test@gmail.com" {
		t.Errorf("repo requestedEmail: expected %q, got %q", "test@gmail.com", fakeUserRepo.requestedEmail)
	}
}

func TestLogin_RepositoryError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	fakeUserRepo := &fakeUserRepository{
		returnedError: dbErr,
	}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Login("test@gmail.com", "password123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, dbErr) {
		t.Errorf("expected %v, got %v", dbErr, err)
	}

	if fakeUserRepo.requestedEmail != "test@gmail.com" {
		t.Errorf("repo requestedEmail: expected %q, got %q", "test@gmail.com", fakeUserRepo.requestedEmail)
	}
}
