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
		t.Error("registerUser should not be called when company does not exist")
	}
}
