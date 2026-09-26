package service

import (
	"errors"
	"task-manager-api/internal/models"
	"testing"
)

type fakeUserRepository struct {
	registeredEmail     string
	registeredPassword  string
	registeredCompanyID int

	registerCalled bool

	requestedEmail string

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

	return models.User{}, f.returnedError
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

func TestRegisterUser_InvalidEmail(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail",
		Password:  "password123",
		CompanyID: 7,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, models.ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}

	if fakeUserRepo.registerCalled {
		t.Error("RegisterUser should not be called if the email is invalid")
	}
}

func TestRegisterUser_InvalidPassword(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "passfls",
		CompanyID: 7,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, models.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}

	if fakeUserRepo.registerCalled {
		t.Error("RegisterUser should not be called if the password is invalid")
	}
}

func TestRegisterUser_InvalidCompanyID(t *testing.T) {
	fakeUserRepo := &fakeUserRepository{}
	fakeCompanyRepo := &fakeCompanyRepository{}
	svc := NewUserService(fakeUserRepo, fakeCompanyRepo)

	_, err := svc.Register(models.RegisterRequest{
		Email:     "test@gmail.com",
		Password:  "password123",
		CompanyID: 0,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if fakeCompanyRepo.ensureCalled {
		t.Error("EnsureExists should not be called if companyID is invalid")
	}

	if !errors.Is(err, models.ErrInvalidCompanyID) {
		t.Errorf("expected ErrInvalidCompanyID, got %v", err)
	}

	if fakeUserRepo.registerCalled {
		t.Error("RegisterUser should not be called if the companyID is invalid")
	}
}
