package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"task-manager-api/internal/models"
	"task-manager-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo        *repository.UserRepository
	companyRepo *repository.CompanyRepository
}

func NewUserService(repo *repository.UserRepository, companyRepo *repository.CompanyRepository) *UserService {
	return &UserService{
		repo:        repo,
		companyRepo: companyRepo,
	}
}

func (s *UserService) Register(user models.RegisterRequest) (models.UserResponse, error) {
	var resp models.UserResponse

	email := strings.TrimSpace(user.Email)

	if len(email) == 0 || len(email) > 255 {
		return models.UserResponse{}, models.ErrInvalidEmail
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return models.UserResponse{}, models.ErrInvalidEmail
	}

	if len(user.Password) < 8 || len(user.Password) > 72 {
		return models.UserResponse{}, models.ErrInvalidPassword
	}

	if user.CompanyID <= 0 {
		return models.UserResponse{}, models.ErrInvalidCompanyID
	}

	if err := s.companyRepo.EnsureExists(user.CompanyID); err != nil {
		return models.UserResponse{}, fmt.Errorf("validate company: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, err
	}
	hashedPassword := string(hashed)

	id, err := s.repo.RegisterUser(email, hashedPassword, user.CompanyID)
	if err != nil {
		return models.UserResponse{}, err
	}

	resp.ID = id
	resp.Email = email
	resp.CompanyID = user.CompanyID
	return resp, nil
}

func (s *UserService) Login(email, password string) (models.UserResponse, error) {
	var resp models.UserResponse

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.UserResponse{}, models.ErrInvalidCredentials
		}

		return models.UserResponse{}, fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return models.UserResponse{}, models.ErrInvalidCredentials
	}

	resp.ID = user.ID
	resp.Email = user.Email
	return resp, nil
}

func (s *UserService) Me(email string) (models.UserResponse, error) {
	var resp models.UserResponse

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.UserResponse{}, models.ErrUserNotFound
		}

		return models.UserResponse{}, fmt.Errorf("get user: %w", err)
	}

	resp.ID = user.ID
	resp.Email = user.Email
	resp.CompanyID = user.CompanyID
	return resp, nil
}
